from mongoengine import (
    BooleanField,
    DateTimeField,
    Document,
    EmailField,
    ListField,
    ReferenceField,
    StringField,
)


class User(Document):
    email = EmailField(unique=True, required=True)
    name = StringField(max_length=255, required=True)
    is_active = BooleanField(default=True)
    is_staff = BooleanField(default=False)

    created_at = DateTimeField()
    updated_at = DateTimeField()

    groups = ListField(StringField(), default=[])  # Relacione IDs de grupos
    user_permissions = ListField(
        StringField(), default=[]
    )  # Relacione IDs de permissões

    meta = {"indexes": [{"fields": ["email"]}]}


class DeviceToken(Document):
    device = ReferenceField("Device", required=True)
    token = StringField(max_length=255, unique=True, required=True)
    is_revoked = BooleanField(default=False)

    created_at = DateTimeField()
    expires_at = DateTimeField()

    meta = {"indexes": [{"fields": ["token"]}, {"fields": ["device", "is_revoked"]}]}
