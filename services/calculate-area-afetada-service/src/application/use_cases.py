import asyncio
from datetime import datetime, timezone
from uuid import UUID
from psycopg import AsyncConnection
from core.affected_area import AffectedArea, ReportPoint, ReportType, assign_report
from core.rules import AreaRules
from infrastructure import database

# data esperado: {"report_id", "user_id", "type", "latitude", "longitude", "datetime"}
async def handle_report_validated(
    conn: AsyncConnection, data: dict, rules: AreaRules, now: datetime
) -> None:
    report = ReportPoint(
        report_id=UUID(data["report_id"]),
        user_id=UUID(data["user_id"]),
        type=ReportType(data["type"]),
        latitude=float(data["latitude"]),
        longitude=float(data["longitude"]),
        occurred_at=_parse_datetime(data["datetime"]),
    )
    # Relato já usado em alguma área (mesmo que ela já tenha fechado) não conta de novo
    if await database.report_exists(conn, report.report_id):
        return

    open_areas = await database.find_open_areas_by_type(conn, report.type, rules)
    area = assign_report(report, open_areas, rules, now)
    await _save(conn, area)

# data esperado: {"area_id", "user_id", "still_happening"}
async def handle_vote(conn: AsyncConnection, data: dict, rules: AreaRules, now: datetime) -> None:
    area = await database.find_area_by_id(conn, UUID(data["area_id"]), rules)
    if area is None:
        return

    area.register_vote(UUID(data["user_id"]), bool(data["still_happening"]), now)
    await _save(conn, area)

# Expira áreas paradas e pergunta se as ativas ainda estão ocorrendo
async def check_areas(conn: AsyncConnection, rules: AreaRules, now: datetime) -> None:
    for area in await database.find_open_areas(conn, rules):
        if area.expire_if_stale(now) or area.request_confirmation(now):
            await _save(conn, area)

async def run_scheduler(conn: AsyncConnection, rules: AreaRules, interval_s: float) -> None:
    while True:
        async with conn.transaction():
            await check_areas(conn, rules, datetime.now(timezone.utc))
        await asyncio.sleep(interval_s)

# Tipo do evento recebido -> caso de uso que trata ele
HANDLERS = {
    "report.validated": handle_report_validated,
    "affected-area.vote": handle_vote,
}

# Salva a área e os eventos dela na mesma transação (padrão Outbox)
async def _save(conn: AsyncConnection, area: AffectedArea) -> None:
    await database.save_area(conn, area)
    for event in area.pull_events():
        await database.outbox_add(conn, event.type.value, _snapshot(area), event.occurred_at)

# Data sem fuso horário é tratada como UTC
def _parse_datetime(value: str) -> datetime:
    parsed = datetime.fromisoformat(value)
    return parsed if parsed.tzinfo else parsed.replace(tzinfo=timezone.utc)

def _snapshot(area: AffectedArea) -> dict:
    center_lat, center_lng = area.center
    return {
        "area_id": str(area.id),
        "type": area.type.value,
        "status": area.status.value,
        "area_m2": round(area.area_m2, 2),
        "center": {"latitude": center_lat, "longitude": center_lng},
        "polygon": area.geojson,
        "reports": len(area.reports),
        "users": area.distinct_users,
    }
