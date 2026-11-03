from collections import defaultdict
from datetime import datetime
from pathlib import Path
from uuid import UUID, uuid4
from psycopg import AsyncConnection
from psycopg.rows import dict_row
from psycopg.types.json import Jsonb
from core.affected_area import AffectedArea, AreaStatus, ReportPoint, ReportType
from core.rules import AreaRules

SCHEMA_PATH = Path(__file__).parent / "schema.sql"
OPEN_STATUSES = [AreaStatus.CANDIDATE.value, AreaStatus.ACTIVE.value]
MAX_INBOX_RETRIES = 5


# autocommit + conn.transaction() explícito: cada bloco de transação fica claro no código
async def connect(database_url: str) -> AsyncConnection:
    return await AsyncConnection.connect(database_url, autocommit=True)

async def create_tables(conn: AsyncConnection) -> None:
    await conn.execute(SCHEMA_PATH.read_text(encoding="utf-8"))


async def find_open_areas_by_type(
    conn: AsyncConnection, type: ReportType, rules: AreaRules
) -> list[AffectedArea]:
    return await _find_areas(conn, rules, "type = %s AND status = ANY(%s)", (type.value, OPEN_STATUSES))

async def find_open_areas(conn: AsyncConnection, rules: AreaRules) -> list[AffectedArea]:
    return await _find_areas(conn, rules, "status = ANY(%s)", (OPEN_STATUSES,))

async def find_area_by_id(conn: AsyncConnection, area_id: UUID, rules: AreaRules) -> AffectedArea | None:
    areas = await _find_areas(conn, rules, "id = %s", (area_id,))
    return areas[0] if areas else None

async def report_exists(conn: AsyncConnection, report_id: UUID) -> bool:
    cur = await conn.execute("SELECT 1 FROM area_reports WHERE report_id = %s", (report_id,))
    return await cur.fetchone() is not None

async def save_area(conn: AsyncConnection, area: AffectedArea) -> None:
    center_lat, center_lng = area.center
    await conn.execute(
        """
        INSERT INTO affected_areas (
            id, type, status, polygon, area_m2, center_lat, center_lng,
            created_at, last_activity_at, last_confirmation_request_at
        )
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (id) DO UPDATE SET
            status = EXCLUDED.status,
            polygon = EXCLUDED.polygon,
            area_m2 = EXCLUDED.area_m2,
            center_lat = EXCLUDED.center_lat,
            center_lng = EXCLUDED.center_lng,
            last_activity_at = EXCLUDED.last_activity_at,
            last_confirmation_request_at = EXCLUDED.last_confirmation_request_at
        """,
        (
            area.id, area.type.value, area.status.value, Jsonb(area.geojson), area.area_m2,
            center_lat, center_lng, area.created_at, area.last_activity_at,
            area.last_confirmation_request_at,
        ),
    )

    async with conn.cursor() as cur:
        await cur.executemany(
            """
            INSERT INTO area_reports (report_id, area_id, user_id, latitude, longitude, occurred_at)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT (report_id) DO NOTHING
            """,
            [
                (r.report_id, area.id, r.user_id, r.latitude, r.longitude, r.occurred_at)
                for r in area.reports
            ],
        )
        if area.votes:
            # Só atualiza o voto se o usuário mudou de resposta
            await cur.executemany(
                """
                INSERT INTO area_votes (area_id, user_id, still_happening)
                VALUES (%s, %s, %s)
                ON CONFLICT (area_id, user_id) DO UPDATE SET
                    still_happening = EXCLUDED.still_happening,
                    voted_at = now()
                WHERE area_votes.still_happening <> EXCLUDED.still_happening
                """,
                [(area.id, user_id, vote) for user_id, vote in area.votes.items()],
            )

async def _find_areas(conn: AsyncConnection, rules: AreaRules, where: str, params: tuple) -> list[AffectedArea]:
    async with conn.cursor(row_factory=dict_row) as cur:
        # FOR UPDATE trava as áreas até o fim da transação, assim duas partes do serviço
        # não mexem na mesma área ao mesmo tempo. ORDER BY id evita deadlock.
        await cur.execute(
            f"SELECT * FROM affected_areas WHERE {where} ORDER BY id FOR UPDATE", params
        )
        area_rows = await cur.fetchall()
        if not area_rows:
            return []

        area_ids = [row["id"] for row in area_rows]
        await cur.execute(
            "SELECT * FROM area_reports WHERE area_id = ANY(%s) ORDER BY occurred_at", (area_ids,)
        )
        report_rows = await cur.fetchall()
        await cur.execute(
            "SELECT area_id, user_id, still_happening FROM area_votes WHERE area_id = ANY(%s)",
            (area_ids,),
        )
        vote_rows = await cur.fetchall()

    reports_by_area = defaultdict(list)
    for row in report_rows:
        reports_by_area[row["area_id"]].append(row)

    votes_by_area = defaultdict(dict)
    for row in vote_rows:
        votes_by_area[row["area_id"]][row["user_id"]] = row["still_happening"]
    return [_to_area(row, reports_by_area[row["id"]], votes_by_area[row["id"]], rules) for row in area_rows]

def _to_area(row: dict, report_rows: list[dict], votes: dict[UUID, bool], rules: AreaRules) -> AffectedArea:
    type = ReportType(row["type"])
    reports = [
        ReportPoint(
            report_id=r["report_id"],
            user_id=r["user_id"],
            type=type,
            latitude=r["latitude"],
            longitude=r["longitude"],
            occurred_at=r["occurred_at"],
        )
        for r in report_rows
    ]
    return AffectedArea(
        id=row["id"],
        type=type,
        rules=rules,
        status=AreaStatus(row["status"]),
        reports=reports,
        votes=votes,
        created_at=row["created_at"],
        last_activity_at=row["last_activity_at"],
        last_confirmation_request_at=row["last_confirmation_request_at"],
    )

# Retorna False se o evento já tinha chegado antes (mensagem duplicada)
async def inbox_add(conn: AsyncConnection, event_id: str, event_type: str, payload: dict) -> bool:
    cur = await conn.execute(
        """
        INSERT INTO inbox (event_id, event_type, payload)
        VALUES (%s, %s, %s)
        ON CONFLICT (event_id) DO NOTHING
        """,
        (event_id, event_type, Jsonb(payload)),
    )
    return cur.rowcount == 1

# SKIP LOCKED: se tiver mais de uma instância do serviço, cada uma pega um evento diferente
async def inbox_fetch_next_pending(conn: AsyncConnection) -> dict | None:
    async with conn.cursor(row_factory=dict_row) as cur:
        await cur.execute(
            """
            SELECT event_id, event_type, payload FROM inbox
            WHERE status = 'PENDING'
            ORDER BY received_at
            LIMIT 1
            FOR UPDATE SKIP LOCKED
            """
        )
        return await cur.fetchone()

async def inbox_mark_processed(conn: AsyncConnection, event_id: str) -> None:
    await conn.execute(
        "UPDATE inbox SET status = 'PROCESSED', processed_at = now() WHERE event_id = %s",
        (event_id,),
    )

async def inbox_mark_rejected(conn: AsyncConnection, event_id: str, error: str) -> None:
    await conn.execute(
        "UPDATE inbox SET status = 'REJECTED', error = %s, processed_at = now() WHERE event_id = %s",
        (error, event_id),
    )

# Erro temporário: tenta de novo depois, até MAX_INBOX_RETRIES
async def inbox_mark_failed(conn: AsyncConnection, event_id: str, error: str) -> None:
    await conn.execute(
        """
        UPDATE inbox SET
            retries = retries + 1,
            error = %s,
            status = CASE WHEN retries + 1 >= %s THEN 'REJECTED' ELSE status END
        WHERE event_id = %s
        """,
        (error, MAX_INBOX_RETRIES, event_id),
    )

async def outbox_add(conn: AsyncConnection, event_type: str, payload: dict, created_at: datetime) -> None:
    await conn.execute(
        "INSERT INTO outbox (id, event_type, payload, created_at) VALUES (%s, %s, %s, %s)",
        (uuid4(), event_type, Jsonb(payload), created_at),
    )

async def outbox_fetch_unpublished(conn: AsyncConnection, limit: int) -> list[dict]:
    async with conn.cursor(row_factory=dict_row) as cur:
        await cur.execute(
            """
            SELECT id, event_type, payload, created_at FROM outbox
            WHERE published_at IS NULL
            ORDER BY created_at
            LIMIT %s
            FOR UPDATE SKIP LOCKED
            """,
            (limit,),
        )
        return await cur.fetchall()

async def outbox_mark_published(conn: AsyncConnection, event_id: UUID) -> None:
    await conn.execute("UPDATE outbox SET published_at = now() WHERE id = %s", (event_id,))
