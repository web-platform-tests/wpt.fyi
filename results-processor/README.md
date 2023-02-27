## Basics

The results processor runs on Python 3.9. The entry point is a Flask web server
(`main.py`). In production, gunicorn is used as the WSGI (see `Dockerfile`) and
the container runs as a custom AppEngine Flex instance (see `app.yaml`).

## Getting started

We can create a virtualenv to recreate a setup close to production for daily
development.

```bash
virtualenv env -p python3.9
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

## Managing dependencies

We maintain our direct dependencies in `requirements.in` and use `pip-compile`
from [pip-tools](https://github.com/jazzband/pip-tools) to generate
`requirements.txt` with pinned versions of all direct and transient
dependencies.

Dependabot is used to automatically update `requirements.txt`. To manually
update dependencies, run the following commands:

```bash
pip3.9 install --user pip-tools
python3.9 -m piptools compile requirements.in
```
