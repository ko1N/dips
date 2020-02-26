worker:
- properly distribute work and have a configurable limit of pipelines / cpulimiter
- if a worker is full redirect messages to different workers
- pipelines should have a set of inputs (either variables, resources or files)
- parse variables
- copy files from/to environments/pipelines
- config file to allow/disallow environments and modules
- config should specify the use of a gpu, pipelines with gpu requirements will be forced on workers with a gpu installed

manager:
- api to paginate/list/start/stop pipelines
- progress tracking
- configurable pipeline triggers (rest, watchfolder, etc)
  -> maybe create new services for watchfolders which will just call the rest api (would be more modular)
- have a set of "predefined" pipelines in pipeline folder
  -> be able to deploy a new pipeline (write pipeline file to folder) or start a pipeline by its filename
- give each running pipeline a tracking id and send it to a worker

pipelines:
- specify gpu use
- specify required docker registry for the given image (and provide a way to configure credentials in the manager and send them to the workers for each pipeline)
- how can we handle pipelines which could scale to multiple servers (e.g. blender crowdrender)?
