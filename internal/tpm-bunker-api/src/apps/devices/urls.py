from uuid import UUID

from django.urls import path, register_converter

from .views import DeviceViewSet


class UUIDConverter:
    regex = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"

    def to_python(self, value):
        return UUID(value)

    def to_url(self, value):
        return str(value)


register_converter(UUIDConverter, "uuid")

urlpatterns = [
    path("", DeviceViewSet.as_view({"get": "list", "post": "create"})),
    path(
        "<uuid:uuid>/",
        DeviceViewSet.as_view(
            {
                "get": "retrieve",
                "put": "update",
                "patch": "partial_update",
                "delete": "destroy",
            }
        ),
    ),
]
