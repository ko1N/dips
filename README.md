<p align="center">
<img src="dips.logo.png" alt="mouseplay" title="mouseplay" />
</p>

# dips - distributed pipeline system

dips is a pipelining system which allows for automated execution of simple workflows.

The project consists of two main parts.
- The `worker` (found in the cmd/worker) directory will run the individual pipelines
- The `manager` accepts new jobs through a REST interface and dispatches them to the workers

The worker and manager are simply interconnected via an AMQP.

# running it

A few test pipelines are provided in the `test` folder which can be simply executed by the `cmd/executor` package:

```
cd cmd/executor
go run executor.go --pipeline ../../test/transcoding.pipe
```

This will pull the alpine:latest image and execute the given commands in a dockerized context.

When working with the entire stack it is recommended to start the compose setup, worker and manager individually:
```
cd deployments/development && docker-compose up
cd cmd/worker && go run worker.go
cd cmd/manager && go run manager.go
```

## License

Licensed under GPL3 License, see [LICENSE](LICENSE).

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted for inclusion in the work by you, shall be licensed as above, without any additional terms or conditions.