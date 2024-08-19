# Нуралиев Кадриддин | Avito Backend Bootcamp

## Установка и конфигурация
+ Склонировать репозиторий:
  ```
  git clone https://github.com/NRKA/backend-bootcamp-assignment-2024
  ```
+ Запустить *docker compose* из корневой директории
  ```make
  docker-compose up -d
  ```
## Использование
### Сервис поддерживает 8 эндпоинтов:
+ `GET    /dummyLogin -- (no Auth)`
+ `POST   /login -- (no Auth)`
+ `POST   /register -- (no Auth)`
+ `GET    /house/{id} -- (Auth only)`
+ `POST    /house/{id}/subscribe -- (Auth only)`
+ `POST   /flat/create -- (Auth only)`
+ `POST   /house/create -- (Moderations only)`
+ `POST   /flat/update -- (Moderations only)`

## Запуск тестов
+ В каждом сервисе написаны тесты как и для репозитория, так и для обработчиков. Чтобы запустить тесты из корневой директории нужно  
  ```make
  make test
  ```
+ Перед запуском тестов необходимо сначала поднять базу с помощью docker-compose up, а затем запустить тесты.