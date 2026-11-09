export const HOST = '192.168.0.0'; //aqui precisa ser o endereço do seu ip do seu computador entre nas 
//configurações e pegue o ip e altere aqui para funcionar tem que ter o app expo go no mesmo wi-fi do pc
// login: login-service (porta 81 no docker-compose) · bff: relatos e votos (porta 8090)
export const URLS = { login: `http://${HOST}:81`, bff: `http://${HOST}:8090` };