## Basics

The results processor runs on Python 3.6. The entry point is a Flask web server
(`main.py`). In production, gunicorn is used as the WSGI (see `Dockerfile`) and
the container runs as a custom AppEnging Flex instance (see `app.yaml`).

## Getting started

We can create a virtualenv to recreate a setup close to production for daily
development.

```bash
virtualenv env -p python3.6
. env/bin/activate
pip install -r requirements.txt
```

## Running tests

We strongly recommend you to run tests **outside of** the virtualenv created
above to avoid running into issues caused by nested virtualenvs (`tox` manages
virtualenv itself).


```bash
deactivate  # or in a different shell
tox
```
