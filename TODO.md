worker:
- properly distribute work and have a configurable limit of pipelines / cpulimiter
- if a worker is full redirect messages to different workers
- pipelines should have a set of inputs (either variables, resources or files)
- parse variables
- copy files from/to environments/pipelines https://github.com/docker/cli/blob/master/cli/command/container/cp.go#L186
- config file to allow/disallow environments and modules
- the pipeline should have a set of default env variables (e.g. container id for docker or env name)
- config should specify the use of a gpu, pipelines with gpu requirements will be forced on workers with a gpu installed

manager:
- api to paginate/list/start/stop pipelines
- progress tracking
- configurable pipeline triggers (rest)
  -> maybe create new services for watchfolders which will just call the rest api (would be more modular)
- have a set of "predefined" pipelines in pipeline folder
  -> be able to deploy a new pipeline (write pipeline file to folder) or start a pipeline by its filename
- give each running pipeline a tracking id and send it to a worker

"pipeline-ui":
- easily deployable and configurable pipelines with config system (so users can configure a pipeline and trigger them)

pipelines:
- will be started with name+params+pipeline script
- specify gpu use
- specify required docker registry for the given image (and provide a way to configure credentials in the manager and send them to the workers for each pipeline)
- how can we handle pipelines which could scale to multiple servers (e.g. blender crowdrender)?

structure:
- swagger specs should be generated into /api folder
- /configs for default configs
- add /scripts directory for build/install/analysis
- add /build folder for docker
- /test should contain additional external test apps and test data
