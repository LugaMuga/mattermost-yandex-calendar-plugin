FROM mattermost/mattermost-preview:6.5.0

RUN apt update && apt install -y procps curl gcc nano bash-completion

# Setup develop environment
RUN sed -i '/SiteURL/c\    \"SiteURL\" : \"http://localhost:8065\",         \"EnableLocalMode\": true,' /mm/mattermost/config/config_docker.json \
    && sed -i '/EnableUploads/c\    \"EnableUploads\": true,      \"EnableHealthCheck\": false,' /mm/mattermost/config/config_docker.json

RUN sed -i '/SiteURL/c\    \"SiteURL\" : \"http://localhost:8065\",\n    \"EnableLocalMode\": true,' /mm/mattermost/config/config.json \
    && sed -i '/EnableUploads/c\    \"EnableUploads\": true,' /mm/mattermost/config/config.json \
    && sed -i '/EnableHealthCheck/c\    \"EnableHealthCheck\": false,' /mm/mattermost/config/config.json

# Install Go and Delve
RUN curl -O https://dl.google.com/go/go1.18.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz \
    && export PATH=$PATH:/usr/local/go/bin \
    && export GOPATH="${HOME}/go" \
    && export PATH="$PATH:${GOPATH}/bin" \
    && go install github.com/go-delve/delve/cmd/dlv@v1.8.2

# Expose port for Delve
EXPOSE 2345
