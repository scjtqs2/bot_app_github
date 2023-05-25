# chrome的docker镜像

## docker-compose方式

`docker-compose up -d`

## docker run 方式

`docker run -d -p 4444:4444 --shm-size="2g" selenium/standalone-chrome:4.1.2-20220217`

## 对应的api地址

`http://服务器ip:4444`

### 使用 grid的方式
`docker-compose -f docker-compose-v3-full-grid.yml up -d`

对应的api地址 `http://服务器ip:4444`