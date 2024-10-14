## Basics

The results processor runs on Python 3.11. The entry point is a Flask web server
(`main.py`). In production, gunicorn is used as the WSGI (see `Dockerfile`) and
the container runs as a custom AppEngine Flex instance (see `app.yaml`).

## Getting started

We can create a virtualenv to recreate a setup close to production for daily
development.

```bash
virtualenv env -p python3.11
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
pip3.11 install --user pip-tools
python3.11 -m piptools compile requirements.in
```

## Local debugging

Debugging is disabled both in production and when running locally by default.
To enable debugging when running locally pass `debug=True` to the `app.run()`
call in the last line of
[`main.py`](https://github.com/web-platform-tests/wpt.fyi/blob/main/results-processor/main.py).
