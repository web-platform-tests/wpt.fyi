## Basics

The results processor runs on Python 3.7. The entry point is a Flask web server
(`main.py`). In production, gunicorn is used as the WSGI (see `Dockerfile`) and
the container runs as a custom AppEnging Flex instance (see `app.yaml`).

## Getting started

We can create a virtualenv to recreate a setup close to production for daily
development.

```bash
virtualenv env -p python3.7
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

We pin all dependencies recursively in `requirements.txt`, and use
[`pyup`](../.pyup.yml) to automatically upgrade them. However,
`requirements.txt` is a flat list and does not differentiate direct
dependencies from transitive ones, so when a direct dependency drops a sub
dependency, the sub dependency will still be kept in `requirements.txt`. Unused
dependencies are usually harmless, but they could sometimes lead to conflicts
(e.g. before: A->B, after A->C but C conflicts with B). Therefore, we manually
maintain `requirements-top_level.txt` that has all the **direct** dependencies
in order to regenerate `requirements.txt` when needed; e.g. when you see this
error from `pip install`:

> ERROR: foo has requirement bar, but you'll have baz which is incompatible

Run the following commands in a **fresh** virtualenv to regenerate
`requirements.txt`:

```bash
pip install -r requirements-top_level.txt
pip freeze > requirements.txt
# Workaround for https://github.com/pypa/pip/issues/4022 on Debian/Ubuntu
sed -i '/pkg-resources==0.0.0/d' requirements.txt
```
