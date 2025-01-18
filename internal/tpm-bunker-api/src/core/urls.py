from core import settings
from django.urls import include, path
from drf_spectacular.views import SpectacularAPIView, SpectacularSwaggerView

from .views import VersionView

urlpatterns = [
    path(f"api/{settings.API_MAJOR}/", VersionView.as_view(), name="api-version"),
    path(
        f"api/{settings.API_MAJOR}/schema/", SpectacularAPIView.as_view(), name="schema"
    ),
    path(
        f"api/{settings.API_MAJOR}/swagger/",
        SpectacularSwaggerView.as_view(url_name="schema"),
        name="swagger",
    ),
    path(f"api/{settings.API_MAJOR}/devices/", include("devices.urls")),
    path(f"api/{settings.API_MAJOR}/auth/", include("authentications.urls")),
    path(f"api/{settings.API_MAJOR}/operations/", include("operations.urls")),
]
