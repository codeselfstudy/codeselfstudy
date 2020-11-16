# TODO: be more specific about the python version.
# Can we shrink it by using alpine linux?
FROM python:3.9
ENV PYTHONUNBUFFERED=1
RUN mkdir -p /app/requirements
WORKDIR /app
COPY requirements/development.txt /app/requirements
RUN pip install -r requirements/development.txt
COPY . /app/
