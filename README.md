Build:
  make

Run:
  ./httpd

Feed:
  curl -F data='{ "key1": "data1", "key2": "data2" }' http://localhost:8080/set  
  for i in $(seq 3 1000) ; do curl -F data="{ \"key$i\": \"data$i\" }" http://localhost:8080/set ; done

Json Get:
  curl http://localhost:8080/get
  
HTML Get:
  curl http://localhost:8080/get/html

CSV Get:
  curl http://localhost:8080/get/csv
