##### Above code was for a job interview. 

To run test:
```bash

docker build -f ./weatherservice/Dockerfile.tests .
```

To run all services
```bash

docker-compose up -d 

```


##### TODO

- [ ] Refactor repetitive code for the three API endpoints in the weather service
- [ ] Refactor the tests, and test returned results (currently returned http code are tested)
- [ ] Add redis for caching dates to avoid hitting the windspeed and temperature API's too much. 
- [ ] Add better documentation explain the purpose of function/structs/interfaces

