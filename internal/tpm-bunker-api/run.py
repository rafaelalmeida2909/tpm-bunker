import argparse
import json
from os import environ, system
from pathlib import Path
from sys import exit

from src.apps.authentications.models import DeviceToken
from src.apps.devices.models import Device
from src.apps.operations.models import EncryptedPackage, Operation, OperationLog

BASE_DIR = Path(__file__).resolve().parent

MODEL_MAPPING = {
    "Device": Device,
    "DeviceToken": DeviceToken,
    "Operation": Operation,
    "EncryptedPackage": EncryptedPackage,
    "OperationLog": OperationLog,
}


def run_command(command: str) -> int:
    """
    Runs a system command and returns the status code.

    Args:
        command (str): The command to run.

    Returns:
        int: The exit status of the command (0 for success, non-zero for failure).
    """
    print(f"\nRunning command: {command}")
    return system(command)


def create_fixtures():
    """
    Carrega os dados de fixtures no MongoDB usando MongoEngine.
    """
    fixtures_path = Path(BASE_DIR, "src", "apps")
    fixtures = ["devices.json", "auths.json"]

    for fixture_name in fixtures:
        found_fixture = False
        for app_path in fixtures_path.iterdir():
            if app_path.is_dir():
                fixture_path = app_path / "fixtures" / fixture_name
                if fixture_path.exists():
                    found_fixture = True
                    with open(fixture_path, "r") as f:
                        data = json.load(f)
                        for record in data:
                            model_name = record["model"].split(".")[-1]
                            model_class = MODEL_MAPPING.get(model_name)
                            if model_class:
                                try:
                                    instance = model_class(**record["fields"])
                                    instance.save()
                                    print(
                                        f"Successfully saved: {model_name} -> {record['fields']}"
                                    )
                                except Exception as e:
                                    print(
                                        f"Failed to save {model_name}: {record['fields']}, Error: {e}"
                                    )
                    print(f"Loaded fixture: {fixture_name} from {fixture_path}")
                    break

        if not found_fixture:
            print(f"Fixture not found: {fixture_name}")


def start_server(port: int) -> None:
    """
    Starts the server based on the provided command-line arguments.
    """
    runserver_command = f'python "{path}" runserver 0.0.0.0:{port}'
    run_command(runserver_command)


def main() -> None:
    """
    Main entry point of the script. Handles database migrations, testing, and
    starting the server based on the command-line arguments.
    """
    print("Skipping migrations because MongoDB is being used.")

    if args.tests:
        test_command = f'python "{path}" test --parallel --noinput'

        if run_command(test_command) != 0:
            exit(1)

    # create_fixtures()

    start_server(args.port)


if __name__ == "__main__":
    HTTPS = environ.get("HTTPS", None)
    path = f"{BASE_DIR}/src/manage.py"

    parser = argparse.ArgumentParser()

    parser.add_argument(
        "-t", "--tests", help="Run tests", dest="tests", action="store_true"
    )

    parser.add_argument(
        "-p",
        "--port",
        help="Port to run the server",
        type=int,
        default=int(environ.get("PORT", 8003)),
    )

    args = parser.parse_args()

    try:
        main()
    except:
        exit(0)
