from datetime import datetime

from rest_framework.serializers import (
    BooleanField,
    CharField,
    DateTimeField,
    Serializer,
    UUIDField,
    ValidationError,
)

from .models import Device


class DeviceRegistrationSerializer(Serializer):
    uuid = UUIDField(help_text="UUID único do dispositivo")
    ek_certificate = CharField(
        help_text="Certificado EK (TPM Endorsement Key) do dispositivo"
    )
    aik = CharField(help_text="Attestation Identity Key do dispositivo")
    public_key = CharField(help_text="Chave pública do dispositivo")

    def validate_uuid(self, value):
        if Device.objects(uuid=value).count() > 0:
            raise ValidationError(
                "A device with this UUID already exists.", code="unique"
            )
        return value

    def create(self, validated_data):
        device = Device.objects.create(**validated_data, registered_at=datetime.now())
        device.save()
        return device

    def update(self, instance, validated_data):
        for attr, value in validated_data.items():
            setattr(instance, attr, value)
        instance.save()
        return instance

    def delete(self, instance):
        instance.delete()


class DeviceSerializer(DeviceRegistrationSerializer):
    is_active = BooleanField(help_text="Indica se o dispositivo está ativo")
    registered_at = DateTimeField(
        read_only=True, help_text="Data de registro do dispositivo"
    )
    last_access = DateTimeField(
        read_only=True, help_text="Último acesso do dispositivo"
    )
