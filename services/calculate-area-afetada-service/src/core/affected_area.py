from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from uuid import UUID, uuid4
from shapely.geometry import Polygon
from core import geometry
from core.rules import AreaRules

class ReportType(str, Enum):
    LANDSLIDE = "Landslide"
    FLOODING = "Flooding"
    THUNDERSTORM = "Thunderstorm"
    FIRE_DISASTER = "Fire Disaster"

class AreaStatus(str, Enum):
    CANDIDATE = "candidate"  # Ainda não aparece no mapa
    ACTIVE = "active"
    RESOLVED = "resolved"
    EXPIRED = "expired"

class AreaEventType(str, Enum):
    # Os valores servem como routing key no RabbitMQ
    DETECTED = "affected-area.detected"
    UPDATED = "affected-area.updated"
    CONFIRMATION_REQUESTED = "affected-area.confirmation-requested"
    RESOLVED = "affected-area.resolved"
    EXPIRED = "affected-area.expired"


@dataclass(frozen=True)
class AreaEvent:
    type: AreaEventType
    area_id: UUID
    occurred_at: datetime

# Relato recebido do report-service, só com o que importa para calcular a área
@dataclass(frozen=True)
class ReportPoint:
    report_id: UUID
    user_id: UUID
    type: ReportType
    latitude: float
    longitude: float
    occurred_at: datetime
    def __post_init__(self):
        if not (-90 <= self.latitude <= 90 and -180 <= self.longitude <= 180):
            print("Latitude ou longitude inválidas")
            return


class AffectedArea:
    def __init__(
        self,
        id: UUID,
        type: ReportType,
        rules: AreaRules,
        status: AreaStatus,
        reports: list[ReportPoint],
        votes: dict[UUID, bool],
        created_at: datetime,
        last_activity_at: datetime,
        last_confirmation_request_at: datetime | None = None,
    ):
        # Usado para reconstruir a área vinda do banco. Para criar uma nova, use start()
        self.id = id
        self.type = type
        self.rules = rules
        self.status = status
        self.reports = reports
        self.votes = votes  # user_id -> True (ainda ocorre) / False (não ocorre mais)
        self.created_at = created_at
        self.last_activity_at = last_activity_at
        self.last_confirmation_request_at = last_confirmation_request_at
        self._events: list[AreaEvent] = []
        self._polygon: Polygon | None = None

    @classmethod
    def start(cls, first_report: ReportPoint, rules: AreaRules, now: datetime) -> "AffectedArea":
        area = cls(
            id=uuid4(),
            type=first_report.type,
            rules=rules,
            status=AreaStatus.CANDIDATE,
            reports=[],
            votes={},
            created_at=now,
            last_activity_at=now,
        )
        area.add_report(first_report, now)
        return area

    @property
    def is_open(self) -> bool:
        return self.status in (AreaStatus.CANDIDATE, AreaStatus.ACTIVE)

    @property
    def distinct_users(self) -> int:
        return len({r.user_id for r in self.reports})

    @property
    def last_report_at(self) -> datetime:
        return max(r.occurred_at for r in self.reports)

    @property
    def polygon_m(self) -> Polygon:
        if self._polygon is None:
            points = [(r.latitude, r.longitude) for r in self.reports]
            self._polygon = geometry.build_polygon(points, self.rules.area_margin_m)
        return self._polygon

    @property
    def area_m2(self) -> float:
        return self.polygon_m.area

    @property
    def center(self) -> tuple[float, float]:
        return geometry.center_latlng(self.polygon_m)

    @property
    def geojson(self) -> dict:
        return geometry.to_latlng_geojson(self.polygon_m)

    def contains(self, report_id: UUID) -> bool:
        return any(r.report_id == report_id for r in self.reports)

    def distance_to(self, report: ReportPoint) -> float:
        return min(
            geometry.distance_m(r.latitude, r.longitude, report.latitude, report.longitude)
            for r in self.reports
        )

    # O relato entra no grupo se for do mesmo tipo, recente e perto de algum relato do grupo
    def accepts(self, report: ReportPoint) -> bool:
        if not self.is_open or report.type != self.type or self.contains(report.report_id):
            return False
        if not self.reports:
            return True
        if abs(report.occurred_at - self.last_report_at) > self.rules.report_time_window:
            return False
        return self.distance_to(report) <= self.rules.max_distance_m

    def add_report(self, report: ReportPoint, now: datetime) -> None:
        if not self.accepts(report):
           print(f"Relato {report.report_id} não pertence à área {self.id}")
           return

        self.reports.append(report)
        self.last_activity_at = now
        self._polygon = None
        # Com relatos suficientes de usuários diferentes, o candidato vira área afetada
        if self.status == AreaStatus.CANDIDATE and self.distinct_users >= self.rules.min_reports:
            self.status = AreaStatus.ACTIVE
            self._record(AreaEventType.DETECTED, now)
        elif self.status == AreaStatus.ACTIVE:
            self._record(AreaEventType.UPDATED, now)

    # Chamado periodicamente para perguntar aos usuários se ainda está ocorrendo
    def request_confirmation(self, now: datetime) -> bool:
        if self.status != AreaStatus.ACTIVE:
            return False
        last = self.last_confirmation_request_at
        if last is not None and now - last < self.rules.confirmation_interval:
            return False

        self.last_confirmation_request_at = now
        self._record(AreaEventType.CONFIRMATION_REQUESTED, now)
        return True

    def register_vote(self, user_id: UUID, still_happening: bool, now: datetime) -> None:
        if self.status != AreaStatus.ACTIVE:
            return

        # Vale só a resposta mais recente de cada usuário
        self.votes[user_id] = still_happening
        if still_happening:
            self.last_activity_at = now  # um "sim" adia a expiração

        # Desmarca a área quando os "não" atingem o mínimo e superam os "sim"
        no_votes = sum(1 for v in self.votes.values() if not v)
        yes_votes = len(self.votes) - no_votes
        if no_votes >= self.rules.min_votes_to_resolve and no_votes > yes_votes:
            self.status = AreaStatus.RESOLVED
            self._record(AreaEventType.RESOLVED, now)

    def expire_if_stale(self, now: datetime) -> bool:
        if not self.is_open or now - self.last_activity_at < self.rules.expiration_after:
            return False

        was_active = self.status == AreaStatus.ACTIVE
        self.status = AreaStatus.EXPIRED
        # Candidato nunca foi publicado, então não precisa avisar ninguém
        if was_active:
            self._record(AreaEventType.EXPIRED, now)
        return True

    # Entrega os eventos pendentes (que vão para o Outbox) e limpa a lista
    def pull_events(self) -> list[AreaEvent]:
        events, self._events = self._events, []
        return events

    def _record(self, type: AreaEventType, now: datetime) -> None:
        self._events.append(AreaEvent(type=type, area_id=self.id, occurred_at=now))


def assign_report(
    report: ReportPoint, open_areas: list[AffectedArea], rules: AreaRules, now: datetime
) -> AffectedArea:
    # Relato repetido não conta de novo
    for area in open_areas:
        if area.contains(report.report_id):
            return area

    # Entra na área compatível mais próxima, ou começa uma nova
    compatible = [a for a in open_areas if a.accepts(report)]
    if not compatible:
        return AffectedArea.start(report, rules, now)

    closest = min(compatible, key=lambda a: a.distance_to(report))
    closest.add_report(report, now)
    return closest
