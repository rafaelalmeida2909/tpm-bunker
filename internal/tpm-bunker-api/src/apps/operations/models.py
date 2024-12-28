from datetime import datetime

from mongoengine import (
    BinaryField,
    DateTimeField,
    DictField,
    Document,
    ReferenceField,
    StringField,
)


class Operation(Document):
    device = ReferenceField("Device", required=True)

    operation_type = StringField(max_length=20, required=True)
    status = StringField(max_length=20, default="PENDING")
    error_message = StringField(null=True, blank=True)

    created_at = DateTimeField(default=datetime.now())

    meta = {
        "indexes": [
            {"fields": ["device", "operation_type"]},
            {"fields": ["status", "created_at"]},
        ]
    }


class EncryptedPackage(Document):
    operation = ReferenceField("Operation", unique=True, required=True)

    encrypted_data = BinaryField(required=True)
    encrypted_symmetric_key = BinaryField(required=True)
    digital_signature = StringField(required=True)
    hash_original = StringField(required=True)  # Hash dos dados originais
    metadata = DictField(default=dict)  # Metadados adicionais

    created_at = DateTimeField(default=datetime.now())

    meta = {"indexes": [{"fields": ["operation", "created_at"]}]}


class OperationLog(Document):
    operation = ReferenceField("Operation", required=True)
    action = StringField(max_length=100, required=True)
    details = DictField(default=dict)
    timestamp = DateTimeField(default=datetime.now())

    meta = {"indexes": [{"fields": ["operation", "timestamp"]}]}
