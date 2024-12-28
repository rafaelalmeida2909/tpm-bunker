# TPM-BUNKER API

To run in docker: `docker-compose up --build` <br>
To erase docker database data: `docker-compose down -v` <br>
To run localhost: `python run.py` <br>

# Test Coverage

To create an "htmlcov" folder with the report, run:

```sh
coverage run src/manage.py test
coverage html
```

To access the coverage report, go to `htmlcov/index.html`
