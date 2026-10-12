from dataclasses import dataclass
from datetime import timedelta

@dataclass(frozen=True)
class AreaRules:
    min_reports: int = 3
    max_distance_m: float = 500.0
    report_time_window: timedelta = timedelta(hours=6)
    expiration_after: timedelta = timedelta(hours=24)
    # 10 segundos só para testes, em produção a ideia é 10 minutos
    confirmation_interval: timedelta = timedelta(seconds=10)
    min_votes_to_resolve: int = 3
    area_margin_m: float = 100.0
