while [ $(curl -s -o /dev/null -w %{http_code} http://localhost:8083/connectors) -ne 200 ] ; do
  echo -e "\t" $(date) " Kafka Connect listener HTTP state: " $(curl -s -o /dev/null -w %{http_code} http://localhost:8083/connectors) " (waiting for 200)"
  sleep 5
done

echo -e $(date) "\n\n--------------\n\o/ Kafka Connect is ready! Listener HTTP state: " $(curl -s -o /dev/null -w %{http_code} http://localhost:8083/connectors) "\n--------------\n"

curl -i -X PUT http://localhost:8083/connectors/film_source/config -H "Content-Type: application/json" -d '{
           "connector.class":"io.confluent.connect.jdbc.JdbcSourceConnector",
           "connection.url":"jdbc:mysql://kafka-legacydb:3306/sakila",
           "connection.user":"admin",
           "connection.password":"admin",
           "poll.interval.ms":"1000",
           "mode":"timestamp",
           "mode":"timestamp",
           "incrementing.column.name": "film_id",
           "timestamp.column.name": "last_update",
           "topic.prefix":"p_film",
           "validate.non.null":"false",
           "transforms": "Cast",
           "transforms.Cast.type": "org.apache.kafka.connect.transforms.Cast$Value",
           "transforms.Cast.spec": "release_year:string",
           "query":"SELECT film_id, title, last_update, language_id, release_year, uuid FROM film"
       }';
