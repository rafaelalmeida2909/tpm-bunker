from datetime import datetime

from rest_framework.serializers import ValidationError

from .models import Device
from .serializers import DeviceSerializer


class DevicesService:
    def register_device(self, serializer_data):
        # Validar certificado EK
        if not self.validate_ek_certificate(serializer_data["ek_certificate"]):
            raise ValidationError({"ek_certificate": "Certificado EK inválido"})

        # Criar e retornar o dispositivo
        device = Device.objects.create(
            uuid=serializer_data["uuid"],
            ek_certificate=serializer_data["ek_certificate"],
            aik=serializer_data["aik"],
            public_key=serializer_data["public_key"],
            registered_at=datetime.now(),
        )

        return DeviceSerializer(device).data

    def validate_ek_certificate(self, ek_certificate):
        # Implementar validação do certificado EK
        # Por enquanto retorna True
        return True
