from devices.models import Device
from django.core.exceptions import PermissionDenied


class DeviceAuthenticationMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        device_uuid = request.headers.get("X-Device-UUID")
        if device_uuid:
            try:
                request.device = Device.objects.get(uuid=device_uuid, is_active=True)
            except Device.DoesNotExist:
                raise PermissionDenied("Dispositivo não encontrado ou inativo")
        return self.get_response(request)
