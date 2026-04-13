docker run --name user_service_db -d \
-e POSTGRES_USER=user \
-e POSTGRES_PASSWORD=password \
-e POSTGRES_DB=user_service_db \
-p 5432:5432 postgres;

docker run --name product_service_db -d \
-e POSTGRES_USER=user \
-e POSTGRES_PASSWORD=password \
-e POSTGRES_DB=product_service_db \
-p 5431:5432 postgres;

docker run -d --name order_service_db \
-e POSTGRES_USER=user \
-e POSTGRES_PASSWORD=password \
-e POSTGRES_DB=order_service_db \
-p 5430:5432 postgres;
