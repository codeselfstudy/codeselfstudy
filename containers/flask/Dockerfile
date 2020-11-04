# BEGIN PANDOC SECTION
# Base #########################################################################
FROM alpine:3.12 AS alpine-builder-base
WORKDIR /app

RUN apk --no-cache add \
        alpine-sdk \
        bash \
        ca-certificates \
        cabal \
        fakeroot \
        ghc \
        git \
        gmp-dev \
        lua5.3-dev \
        pkgconfig \
        zlib-dev

COPY cabal.root.config /root/.cabal/config
RUN cabal --version \
    && ghc --version \
    && cabal new-update

# Builder ######################################################################
FROM alpine-builder-base as alpine-builder
ARG pandoc_commit=master
RUN git clone --branch=$pandoc_commit --depth=1 --quiet \
    https://github.com/jgm/pandoc /usr/src/pandoc

COPY ./docker/pandoc-$pandoc_commit.project.freeze \
     /usr/src/pandoc/cabal.project.freeze

# Copy placeholder for pandoc-citeproc. This might be overwritten with a
# proper binary later.
COPY ./docker/placeholder-pandoc-citeproc.sh \
    /usr/local/bin/pandoc-citeproc

# Install Haskell dependencies
WORKDIR /usr/src/pandoc
# Add pandoc-crossref to project
ARG without_crossref=
RUN test -n "$without_crossref" || \
    printf "extra-packages: pandoc-crossref\n" > cabal.project.local;

# Additional projects to compile alongside pandoc
ARG extra_packages="pandoc-citeproc pandoc-crossref"

RUN cabal new-update \
    && cabal new-build \
        --disable-tests \
        --jobs \
        . $extra_packages

# Cabal's exec stripping doesn't seem to work reliably, let's do it here.
RUN find dist-newstyle \
         -name 'pandoc*' -type f -perm -u+x \
         -exec strip '{}' ';' \
         -exec cp '{}' /usr/local/bin/ ';'

# Core #########################################################################
FROM alpine:3.12 AS alpine-core
ARG pandoc_version=edge
LABEL maintainer='Albert Krewinkel <albert+pandoc@zeitkraut.de>'
LABEL org.pandoc.maintainer='Albert Krewinkel <albert+pandoc@zeitkraut.de>'
LABEL org.pandoc.author "John MacFarlane"
LABEL org.pandoc.version "$pandoc_version"

WORKDIR /data
ENTRYPOINT ["/usr/local/bin/pandoc"]

COPY --from=alpine-builder \
  /usr/local/bin/pandoc \
  /usr/local/bin/pandoc-citeproc \
  /usr/local/bin/

# Reinstall any system packages required for runtime.
RUN apk --no-cache add \
        gmp \
        libffi \
        lua5.3 \
        lua5.3-lpeg

# Crossref #####################################################################
# FROM alpine-core AS alpine-crossref
FROM tiangolo/uwsgi-nginx-flask:python3.6-alpine3.7 AS alpine-crossref
COPY --from=alpine-builder \
  /usr/local/bin/pandoc-crossref \
  /usr/local/bin/
# END PANDOC SECTION

RUN apk --update add bash vim
ENV STATIC_URL /static
ENV STATIC_PATH /app/app/static
COPY ./requirements.txt /var/www/requirements.txt
RUN pip install -r /var/www/requirements.txt

# FROM tiangolo/uwsgi-nginx-flask:python3.6-alpine3.7
# RUN apk --update add bash vim
# ENV STATIC_URL /static
# ENV STATIC_PATH /app/app/static
# COPY ./requirements.txt /var/www/requirements.txt
# RUN pip install -r /var/www/requirements.txt
