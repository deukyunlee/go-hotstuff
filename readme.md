# go-hotstuff
The go-hotstuff project is a lightweight implementation of the HotStuff consensus algorithm, designed to provide scalable and efficient distributed consensus for blockchain applications. This protocol is currently under development and is not yet operational. Development is progressing in accordance with the following research paper:

- https://arxiv.org/pdf/1803.05069

## How to Build & Run
### Using Docker
To build and run the application with Docker, use the following command:
    `docker compose up --build`

### Natively
To build and run the application natively, execute:
    `sh start.sh`

## How to Kill process
### Using Docker
To stop and remove the Docker containers, run:
    `docker compose down`

### Natively
To kill the running process natively, use:
    `sh kill.sh`

## Dependencies
- Go (version 1.17 or higher)
- Docker (if using Docker)

## Contributing
Contributions are welcome! If you would like to contribute to the HotStuff Small Protocol, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Commit your changes.
4. Push your branch to your forked repository.
5. Open a pull request.