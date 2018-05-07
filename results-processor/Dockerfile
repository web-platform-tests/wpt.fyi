FROM gcr.io/google-appengine/python

USER root
WORKDIR /root

RUN apt-get update -q
RUN apt-get upgrade -qy
RUN apt-get install -qy curl

RUN curl -s https://sdk.cloud.google.com > install-gcloud.sh
RUN bash install-gcloud.sh --disable-prompts --install-dir=/opt

# This will take effect at run time.
RUN echo 'PATH=$PATH:/opt/google-cloud-sdk/bin' >> .profile
# This takes effect during the build time.
ENV PATH=$PATH:/opt/google-cloud-sdk/bin

RUN gcloud config set disable_usage_reporting false
