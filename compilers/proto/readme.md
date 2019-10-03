docker run -it -v \$(pwd):/opt embly/protoc protoc -o thing -I /opt /opt/data.proto
