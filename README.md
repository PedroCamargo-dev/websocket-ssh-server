# Go WebSocket SSH Server

Este projeto é um servidor WebSocket em Go que permite aos clientes se conectarem a servidores SSH e executarem comandos através de uma interface WebSocket. Ele suporta múltiplos clientes, cada um podendo estabelecer uma conexão SSH com um servidor remoto e executar comandos em uma sessão interativa.

## Funcionalidades

- **Conexões SSH via WebSocket**: Permite que clientes estabeleçam conexões SSH com servidores remotos usando WebSockets.
- **Execução de comandos**: Envia comandos do cliente para o servidor SSH e retorna a saída do comando para o cliente via WebSocket.
- **Gerenciamento de múltiplos clientes**: Suporta múltiplos clientes conectados simultaneamente, cada um com sua própria conexão SSH.
- **Autenticação SSH**: Suporta autenticação via senha ou chave privada SSH.

## Arquitetura

Este servidor é composto por três pacotes principais:

- **services**: Contém a lógica de conexão SSH, gerenciamento de sessões e execução de comandos.
- **handlers**: Gerencia as conexões WebSocket e interage com os serviços SSH.
- **clients**: Gerencia os clientes conectados, mantendo o estado das conexões WebSocket e SSH.

## Requisitos

- Go 1.18 ou superior.
- Bibliotecas externas:
  - `github.com/gorilla/websocket`: Para comunicação via WebSocket.
  - `golang.org/x/crypto/ssh`: Para gerenciar conexões SSH.

## Instalação

1. Clone o repositório:

   ```sh
   git clone https://github.com/PedroCamargo-dev/websocket-ssh-server.git
   cd websocket-ssh-server
   ```

2. Instale as dependências:

   ```sh
   go mod tidy
   ```

3. Compile o projeto:

   ```sh
   go build -o websocket-ssh-server cmd/main.go
   ```

4. Execute a aplicação:

   ```sh
      PORT=8080 ./websocket-ssh-server
      ```

      O servidor WebSocket SSH estará disponível na porta definida pela variável de ambiente `PORT`, que por padrão é 8080.

## Utilização com Docker Compose

1. Clone o repositório:

   ```sh
   git clone https://github.com/PedroCamargo-dev/websocket-ssh-server.git
   cd websocket-ssh-server
   ```

2. Execute o Docker Compose:

   ```sh
   docker-compose up --build
   ```

   O servidor WebSocket SSH estará disponível na porta 8080.

## Uso

1. O servidor aceita conexões WebSocket. Quando um cliente se conecta via WebSocket, ele pode enviar mensagens para configurar a conexão SSH e executar comandos na sessão SSH interativa.

   - **Configuração de conexão SSH**:
     Para iniciar uma conexão SSH, o cliente deve enviar uma mensagem JSON no seguinte formato:

     ```json
     {
       "type": "config",
       "sshData": {
         "host": "example.com",
         "port": "22",
         "username": "user",
         "password": "password", // Ou use "privateKey" se for o método de autenticação
         "authMethod": "password" // "password" ou "privateKey"
       }
     }
     ```

     O servidor irá estabelecer a conexão SSH com o servidor remoto e, se bem-sucedido, começará uma sessão interativa.

   - **Envio de comandos**:
     O cliente pode enviar comandos para o servidor SSH. O formato da mensagem para enviar um comando é o seguinte:

     ```json
     {
       "type": "command",
       "command": "ls -l"
     }
     ```

     A resposta do comando será retornada através do WebSocket.

## Estrutura do Código

### `services`

- **SetupSSHConnection(clientID string, config ConfigMessage)**: Estabelece uma conexão SSH para o cliente com base nas configurações fornecidas.
- **startShell(clientID string)**: Inicia uma sessão de shell interativa via SSH para o cliente.
- **ExecuteCommand(client \*clients.Client, command string)**: Envia um comando para ser executado na sessão SSH.

### `handlers`

- **HandleWebSocket(w http.ResponseWriter, r \*http.Request)**: Gerencia a conexão WebSocket, fazendo o upgrade do protocolo HTTP para WebSocket e processando mensagens de configuração e comando.
- **handleMessage(clientID string, message []byte)**: Processa mensagens de configuração e comando enviadas pelo cliente.

### `clients`

- **GetClient(clientID string)**: Retorna o cliente associado ao `clientID`.
- **AddClient(clientID string, client \*Client)**: Adiciona um cliente ao mapa de clientes.
- **CleanupConnection(clientID string)**: Limpa a conexão do cliente, fechando a conexão SSH e WebSocket.

## Próximos Passos

1. **Melhorar tratativas de erros**

   - Adicionar tratamento de erros para lidar com falhas na conexão SSH.
   - Implementar mecanismos de recuperação para reconectar automaticamente em caso de desconexão.

2. **Implementar Sistema de Login Seguro**:

   - Utilizar OAuth 2.0 ou JWT para autenticação.

3. **Registrar Atividades dos Usuários**:

   - Registrar atividades.
   - Exibir logs de ações, incluindo IP de login e timestamp.
   - Monitorar tentativas de login falhas e bem-sucedidas.
   - Armazenar logs de desconexões e tempo de sessão.

4. **Acessos Simultâneos**:

   - Implementar funcionalidade de sessões simultâneas, permitindo trabalhar em conjunto.
   - Limitar o número de sessões simultâneas por usuário.

5. **Adicionar Chat em Tempo Real Durante a Sessão**:

   - Implementar funcionalidade de chat para comunicação durante a sessão SSH compartilhada.
   - Permitir troca de mensagens entre os participantes da sessão.

6. **Criar Sessões Temporárias de Colaboração**:
   - Permitir criação de sessões temporárias com links exclusivos.
   - Configurar expiração automática das sessões ou fechamento por inatividade.

## Contribuições

Se você deseja contribuir com novas funcionalidades ou correções de bugs, fique à vontade para abrir uma "issue" ou enviar um "pull request". As sugestões acima são algumas das funcionalidades que podem ser adicionadas ao projeto, mas todas as contribuições são bem-vindas.

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Faça push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Este projeto está licenciado sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.
