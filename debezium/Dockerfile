FROM debezium/server:2.4

# Scripting SMT用の依存関係を追加
USER root
RUN cd /debezium/lib && \
    curl -O https://repo1.maven.org/maven2/io/debezium/debezium-scripting/2.4.2.Final/debezium-scripting-2.4.2.Final.jar && \
    curl -O https://repo1.maven.org/maven2/org/codehaus/groovy/groovy/3.0.19/groovy-3.0.19.jar && \
    curl -O https://repo1.maven.org/maven2/org/codehaus/groovy/groovy-json/3.0.19/groovy-json-3.0.19.jar && \
    curl -O https://repo1.maven.org/maven2/org/codehaus/groovy/groovy-jsr223/3.0.19/groovy-jsr223-3.0.19.jar
USER 185
