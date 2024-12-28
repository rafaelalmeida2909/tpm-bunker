from mongoengine import BooleanField, DateTimeField, Document, StringField, UUIDField


class Device(Document):
    uuid = UUIDField(binary=False, unique=True, required=True)
    ek_certificate = StringField(required=True)  # TPM Endorsement Key
    aik = StringField(required=True)  # Attestation Identity Key
    public_key = StringField(required=True)
    is_active = BooleanField(default=True)

    registered_at = DateTimeField()
    last_access = DateTimeField()

    meta = {"indexes": [{"fields": ["uuid"]}]}
