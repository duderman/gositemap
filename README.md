# Importer benchmark

Sitemap importer _-ish_ . Written in Go to figure out if we could benefit from using this language

## Details

After our discussion about the new architecture of importers, I've decided to implement a test importer. Instead of mimicking the existing tooling we have I've implemented a workflow matching our future approach to how all the importers suppose to work. Basically, it should download the data, parse it, extract the required data to TSV and upload it to S3. Keeping that in mind this particular importer does the following steps:

1. Parse provided via env variable payload
2. Download sitemap from specified URL
3. Parse XML
4. Output links to TSV file
5. Upload the file to S3

To properly compare the whole workflow and efficiency the script was written in Ruby as well

## Benchmarks

Test | **Go** | **Ruby** | Difference
--- | --- | --- | ---
Resulting docker image size | 17 MB | 74 MB | _78%_
Image build time | ~ 45 sec | ~ 70 sec | _36%_
Run time (_Average of 10 runs_) | 1163 ms | 2356 ms | _51%_
Startup time (_Average of 10 runs_) | 974 ms | 1538 ms | _37%_
Memory consumption | 70 MB | 200 MB | _65%_

Let's go over those numbers:
1. **Image size** - Scripts written in go are compiled into a small executable file containing everything required for it to run which makes the resulting docker image as small as possible. It only contains OS and the executable itself. The Ruby version, on the other hand, has to be shipped along with the interpreter and all the required libraries which makes the final image much bigger
2. **Image build time** - In this example, Ruby image could be built a bit faster but due to usage of Nokogiri for XML parsing it's compilation time results in a much longer image building time
3. **Run time** - Represents a total time taken by the script to finish. Measured from within the script itself
4. **Startup time** - Measured by the small script called [`test.js`](test.js). It basically shows how much time passed between calling `docker run` command and the start of the script itself. Wanted to measure it to figure out how long it takes to spin up the whole environment. In the case of Ruby, it takes a bit longer because you have to spin up the interpreter and load all the libraries
5. **Memory consumption** - System memory used by the process at the end of the script. Go is much more efficient due to its strongly typed nature I guess

## How to use

Comes with a [Makefile](Makefile) containing all the required commands

To build containers:
```sh
make build
# or separately
make build-rb
make build-go
```

To run the script:
```sh
make run-minio # starts a local s3 attached to a new network
make run-rb
make run-go
```

To remove created network and stop local s3:
```sh
make clean
```

To measure startup time:
```sh
node test.js make run-rb
node test.js make run-go
```

## Comments

Using a different paradigm in development writing code in Golang brings some complexity. Strong typing might be a bit difficult when working with third-party resources and flexible schema. On the other hand, combined with code compilation it gives more certainty eliminating lots of runtime errors. Also it worth to consider the tooling around the language. Go has a lot of really powerful features like formatting, documentation and testing out of the box. Also from my small experience with this language I've noticed how well it supported by editors. I'm using VS Code and it provides a lot more functionality when using Go like better refactoring and testing. And it comes right out of the box as well when with Ruby I had to spend a good amount of time setting everything up for a comfort development process. I guess it's much easier to support stricter languages. On the contrary, the time it took me to implement such a small script in Go is a bit disappointing. But it wasn't that hard to start coding right away tbh.
