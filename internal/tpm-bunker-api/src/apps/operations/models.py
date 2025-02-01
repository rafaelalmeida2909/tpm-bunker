import logging
from datetime import datetime

from gridfs import GridFS
from mongoengine import (
    BinaryField,
    DateTimeField,
    DictField,
    Document,
    FloatField,
    ObjectIdField,
    ReferenceField,
    StringField,
)
from mongoengine.connection import get_db
from rest_framework.serializers import ValidationError


class Operation(Document):
    device = ReferenceField("Device", required=True)

    operation_type = StringField(max_length=20, required=True)
    status = StringField(max_length=20, default="PENDING")
    error_message = StringField(null=True, blank=True)

    created_at = DateTimeField(default=datetime.now)

    meta = {
        "indexes": [
            {"fields": ["device", "operation_type"]},
            {"fields": ["status", "created_at"]},
        ]
    }


class EncryptedPackage(Document):
    operation = ReferenceField("Operation", unique=True, required=True)

    file_name = StringField(required=True)
    file_size = FloatField(required=True)
    encrypted_data_id = ObjectIdField(required=True)
    encrypted_symmetric_key = BinaryField(required=True)
    digital_signature = StringField(required=True)
    hash_original = StringField(required=True)  # Hash dos dados originais
    metadata = DictField(default=dict)  # Metadados adicionais

    created_at = DateTimeField(default=datetime.now)

    meta = {"indexes": [{"fields": ["operation", "created_at"]}]}

    @property
    def encrypted_data(self):
        try:
            db = get_db()
            fs = GridFS(db)

            grid_file = fs.get(self.encrypted_data_id)
            data = grid_file.read()

            return data
        except Exception as e:
            raise ValidationError(f"Erro ao recuperar arquivo do GridFS: {str(e)}")

    @encrypted_data.setter
    def encrypted_data(self, value):
        try:
            db = get_db()
            fs = GridFS(db)

            self.encrypted_data_id = fs.put(value)

        except Exception as e:
            raise ValidationError(f"Erro ao salvar arquivo no GridFS: {str(e)}")


class OperationLog(Document):
    operation = ReferenceField("Operation", required=True)
    action = StringField(max_length=100, required=True)
    details = DictField(default=dict)
    timestamp = DateTimeField(default=datetime.now)

    meta = {"indexes": [{"fields": ["operation", "timestamp"]}]}
