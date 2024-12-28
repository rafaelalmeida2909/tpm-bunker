# project/utils/no_migrations.py


class DisableMigrations:
    def __contains__(self, item):
        return True

    def __getitem__(self, item):
        return None
