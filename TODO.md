worker:
- properly distribute work and have a configurable limit of pipelines / cpulimiter
- if a worker is full redirect messages to different workers
- pipelines should have a set of inputs (either variables, resources or files)
- parse variables
- config file to allow/disallow environments and modules
- the pipeline should have a set of default env variables (e.g. container id for docker or env name)
- config should specify the use of a gpu, pipelines with gpu requirements will be forced on workers with a gpu installed
- buffer tracking messages before sending them to fast
- log file with log rotation
- keep local database (e.g. store docker containers in db so we can clean them up in case of a crash!)

manager:
- api to paginate/list/start/stop pipelines
- progress tracking
- configurable pipeline triggers (rest)
  -> maybe create new services for watchfolders which will just call the rest api (would be more modular)
- have a set of "predefined" pipelines in pipeline folder
  -> be able to deploy a new pipeline (write pipeline file to folder) or start a pipeline by its filename
- give each running pipeline a tracking id and send it to a worker
- list pagination: https://github.com/moehlone/mongodm-example/blob/master/controllers/user.go
- store secrets and send them to the worker (e.g. git credentials, docker login, etc)
- store database object of pipeline and manage them
- start a pipeline with a registered pipeline + just start any job

event-tracking:
- log messages
- console output (stdout + stderr seperatly)
- progress update

"pipeline-ui":
- easily deployable and configurable pipelines with config system (so users can configure a pipeline and trigger them)

pipelines:
- will be started with name+params+pipeline script
- specify gpu use
- specify required docker registry for the given image (and provide a way to configure credentials in the manager and send them to the workers for each pipeline)
- how can we handle pipelines which could scale to multiple servers (e.g. blender crowdrender)?
- properly track pwdh

tasks:
- should be made generic and should be derived/extended in case task is async/has when condition/etc
- properly track pwd
- implement executionResult for other modules than shell
- rename engine.go to runtime.go maybe?
- rename modules package to extensions package
- seperate interfaces from implementations in environment and storage (like we did with extensions)

environment/docker:
- cd does not change pwd

extensions/wget:
- error is not handled properly
- should be installed automatically or give a proper error...

extensions/storage:
- track progress on file tranfers
- ls should store result in a variable and pipe (when using register cmd)
- we need a command to copy the final result file(s) out of the storage before deleting the storage

structure:
- /configs for default configs
- add /scripts directory for build/install/analysis
- add /build folder for docker
- /test should contain additional external test apps and test data
- create cmd for manual pipeline execution (for development of pipelines)
- use error wrapping!!!
- move storage to pkg

pipeline ideas:
- upscale/upsample/ffmpeg transcode
- spotify downloader
- blender render
- 

misc:
- persist amqp + update amqp code
