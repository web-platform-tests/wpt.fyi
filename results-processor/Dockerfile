FROM python:3.11.13-bookworm

# Install runtime dependencies.
# python3-crcmod for faster gsutil checksum
# python3-virtualenv for virtualenv
# https://cloud.google.com/storage/docs/gsutil/commands/rsync#slow-checksums
RUN apt-get update -q && apt-get install -qy python3-crcmod python3-virtualenv  && apt-get clean
# gcloud SDK
RUN curl -s https://sdk.cloud.google.com > install-gcloud.sh
RUN bash install-gcloud.sh --disable-prompts --install-dir=/opt > /dev/null
ENV PATH=/opt/google-cloud-sdk/bin:$PATH
RUN gcloud config set disable_usage_reporting false
# This file caches whether we are running on GCE. When created during the image
# building process, the file says False (because images aren't built on GCE),
# which makes gcloud & gsutil fail to use the default service account in this
# container until the cache expires (~5 minutes).
RUN rm -f $HOME/.config/gcloud/gce

# Setup and activate virtualenv.
RUN virtualenv -p python3.11 /env
ENV VIRTUAL_ENV /env
ENV PATH /env/bin:$PATH

# WORKDIR needs to be set explicitly to /app
WORKDIR /app

# Install Python dependencies.
ADD requirements.txt /app/
RUN pip install -r requirements.txt

ADD . /app/
# The number of workers should always be 2: one for processing tasks, the other
# for responding health checks. Scale the service by increasing the number of
# instances instead.
# The timeout for gunicorn should be significantly longer than the timeout in
# main.py for liveness checks, because when things go wrong we'd like AppEngine
# to restart a fresh Docker instance instead of having gunicorn to restart the
# worker (which would require extra cleanup/recovery logic).
CMD exec gunicorn --bind :$PORT --timeout 7200 --workers 2 main:app
