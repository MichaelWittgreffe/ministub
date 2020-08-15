# ministub
A small API stubbing tool for microservice dependency simulation, allows the developer to define an API schema in YAML as well as follow-on actions allowing two-way communication between microservices. I developed this tool primarily to allow the mocking of other microservices without then having to spiral out and mock theirs etc, as an integration stage for a microservices CI pipeline before deploying to a full test environment and further automated testing. Additionally, to aid in the development of said microservice without having to run all the dependencies on a local machine or a development server.

- Define a YAML file with an API and actions on request
- Define a series of input requests to recieve and return different status codes on different occurances (including random injection)
- Define a series of follow-on subsiquent external requests, also injecting different requests to allow for failure and resilience testing
- Extract metrics either from an API or write a results file on shutdown
- Once all requests/responses have been satisfied, add option to exit the tool