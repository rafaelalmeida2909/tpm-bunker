from main_core.settings import BASE_DIR


def get_fixture_path(apps: list[str]) -> list[str]:
    """
    Generates a list of file paths for fixture files based on the given app names.

    Args:
        apps (list[str]): A list of application names for which to generate fixture
        file paths.

    Returns:
        list[str]: A list of file paths where each path corresponds to a fixture file
                   in the format '<BASE_DIR>/apps/<app>/fixtures/<app>.json'.
    """
    return [f"{BASE_DIR}/apps/{app}/fixtures/{app}.json" for app in apps]
