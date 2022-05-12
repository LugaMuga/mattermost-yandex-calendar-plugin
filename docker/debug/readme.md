### Launch plugin in local Docker for development
1. Run test Mattermost with docker-compose
 > docker-compose up --build --force-recreate -d

2. Go to project root directory (https://developers.mattermost.com/integrate/plugins/developer-setup/)
```bash
    export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
    export MM_ADMIN_USERNAME=${ADMIN_USERNAME}
    export MM_ADMIN_PASSWORD=${ADMIN_PASSWORD}
    export GO_BUILD_FLAGS='-gcflags "all=-N -l"'
    make deploy
```

### Debug plugin in Docker
For debug your plugin in docker container you need to compile plugin with specifics flags and run docker without core security
1. Connect to CLI in docker MM container
 > docker exec -it test_mm /bin/bash

2. Start Delve debugger by attach to plugin process
 > /root/go/bin/dlv attach $(pidof plugins/com.github.lugamuga.mattermost-yandex-calendar-plugin/server/dist/plugin-linux-amd64) --listen :2345 --headless=true --api-version=2 --accept-multiclient

3. Connect to `localhost:2345` from your IDE

For fast debug start you can assign credentials
```
${ADMIN_USERNAME} = admin
${ADMIN_PASSWORD} = password
```
and then execute
```bash
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
export GO_BUILD_FLAGS='-gcflags "all=-N -l"'
make docker-debug-deploy
```