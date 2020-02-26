worker:
- properly distribute work and have a configurable limit of pipelines / cpulimiter
- if a worker is full redirect messages to different workers
- pipelines should have a set of inputs (either variables, resources or files)
- parse variables
- copy files from/to environments/pipelines https://github.com/docker/cli/blob/master/cli/command/container/cp.go#L186
- config file to allow/disallow environments and modules
- the pipeline should have a set of default env variables (e.g. container id for docker or env name)

manager:
- api to paginate/list/start/stop pipelines
- progress tracking
- configurable pipeline triggers (rest, watchfolder, etc)
  -> maybe create new services for watchfolders which will just call the rest api (would be more modular)
- have a set of "predefined" pipelines in pipeline folder
  -> be able to deploy a new pipeline (write pipeline file to folder) or start a pipeline by its filename
