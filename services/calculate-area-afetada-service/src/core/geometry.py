from pyproj import Geod, Transformer
from shapely.geometry import MultiPoint, Polygon, mapping
from shapely.ops import transform

# EPSG:5880 = SIRGAS 2000 / Brazil Polyconic, projeção em metros feita para o Brasil
_TO_METERS = Transformer.from_crs("EPSG:4326", "EPSG:5880", always_xy=True)
_TO_LATLNG = Transformer.from_crs("EPSG:5880", "EPSG:4326", always_xy=True)
_GEOD = Geod(ellps="WGS84")

def distance_m(lat1: float, lng1: float, lat2: float, lng2: float) -> float:
    _, _, dist = _GEOD.inv(lng1, lat1, lng2, lat2)
    return dist

# Contorno que envolve todos os relatos + uma margem ao redor (resultado em metros)
def build_polygon(points: list[tuple[float, float]], margin_m: float) -> Polygon:
    projected = [_TO_METERS.transform(lng, lat) for lat, lng in points]
    return MultiPoint(projected).convex_hull.buffer(margin_m)

# Volta para lat/lng no formato GeoJSON, que é o que o mapa entende
def to_latlng_geojson(polygon_m: Polygon) -> dict:
    return mapping(transform(_TO_LATLNG.transform, polygon_m))

def center_latlng(polygon_m: Polygon) -> tuple[float, float]:
    lng, lat = _TO_LATLNG.transform(polygon_m.centroid.x, polygon_m.centroid.y)
    return lat, lng
