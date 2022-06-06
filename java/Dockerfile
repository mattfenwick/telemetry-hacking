FROM azul/zulu-openjdk-alpine:17-jre


ADD opentelemetry-javaagent.jar opentelemetry-javaagent.jar
ADD /target/hacking-1.0-SNAPSHOT-jar-with-dependencies.jar hacking.jar
ENTRYPOINT ["java", \
            "-javaagent:opentelemetry-javaagent.jar", \
            "-Dotel.metrics.exporter=none", \
            "-Dotel.service.name=middle-java", \
            "-Dotel.traces.port=16686", \
            "-Dotel.traces.exporter=jaeger", \
            "-jar", "hacking.jar"]
