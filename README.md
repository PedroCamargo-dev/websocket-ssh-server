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
   docker compose up --build
   ```

   O servidor WebSocket SSH estará disponível na porta 8080.

## Uso

1. O servidor aceita conexões WebSocket. Quando um cliente se conecta via WebSocket, ele pode enviar mensagens para configurar a conexão SSH e executar comandos na sessão SSH interativa.

   - **Configuração de conexão SSH com senha**:
     Para iniciar uma conexão SSH, o cliente deve enviar uma mensagem JSON no seguinte formato:

     ```json
      {
         "type": "config",
         "content": "{\"host\":\"172.26.207.37\",\"port\":22,\"user\":\"pedrocamargo\", \"password\":\"Senha\",\"privateKey\":\"\"}"
      }
     ```

      O servidor irá estabelecer a conexão SSH com o servidor remoto e, se bem-sucedido, começará uma sessão interativa.  

   - **Configuração de conexão SSH com chave Privada**:
      Para autenticar com uma chave privada SSH, o cliente deve fornecer a chave privada no campo `privateKey` da mensagem de configuração.

     ```json
      {
         "type": "config",
         "content": "{\"host\":\"192.168.1.111\",\"port\":22,\"user\":\"pedrocamargo\",\"password\":\"Senha\",\"privateKey\":\"privateKey\"}"
      }
     ```

      O servidor irá estabelecer a conexão SSH com o servidor remoto e, se bem-sucedido, começará uma sessão interativa.

   - **Envio de comandos**:
     O cliente pode enviar comandos para o servidor SSH. O formato da mensagem para enviar um comando é o seguinte:

     ```json
      {
         "type": "input",
         "content": "ls\r"
      }
     ```

     ***Obs: Você deve adicionar o caractere de nova linha `\r` ao final do comando para indicar o fim da entrada.***

      A resposta do comando será retornada através do WebSocket com o seguinte formato, o `content` contém a saída do comando em ANSI:

     ```json
      {
         "type": "output",
         "content": "test.txt\r\n\u001b[?2004h\u001b]0;pedrocamargo@SandboxUbuntu: ~\u0007\u001b[01;32mpedrocamargo@SandboxUbuntu\u001b[00m:\u001b[01;34m~\u001b[00m$ "
      }
     ```

## Mensagens de Erro

A seguir estão os erros identificados no projeto, junto com seus respectivos códigos e mensagens:

## Mapeamento de Erros no WebSocket

| **Código**              | **Descrição**                                                                                         | **Mensagem ao Cliente**                            |
|-------------------------|-----------------------------------------------------------------------------------------------------|---------------------------------------------------|
| `WS_UPGRADE_FAILED`      | Falha ao realizar o upgrade do protocolo HTTP para WebSocket.                                       | "Failed to upgrade to WebSocket"                 |
| `MISSING_CLIENT_ID`      | Falta do cabeçalho `Sec-WebSocket-Key` na conexão WebSocket.                                        | "Sec-WebSocket-Key not provided"                 |
| `WS_READ_FAILED`         | Erro ao tentar ler uma mensagem do WebSocket.                                                      | "Failed to read WebSocket message"               |
| `SSH_SESSION_FAILED`     | Falha ao iniciar uma sessão SSH com as configurações fornecidas.                                   | "Failed to start SSH session"                    |
| `RESIZE_FAILED`          | Erro ao redimensionar o terminal da sessão SSH.                                                    | "Failed to resize terminal"                      |
| `UNKNOWN_MESSAGE_TYPE`   | Tipo de mensagem desconhecido recebido pelo servidor WebSocket.                                    | "Unknown message type received"                  |

## Mapeamento de Erros no SSH

| **Código do Erro**              | **Descrição**                                                                                          | **Cenário**                                                                                                                                                                                                 | **Mensagem ao Cliente**                            |
|----------------------------------|--------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------|
| `INVALID_CONFIG`                 | Configuração SSH inválida no JSON enviado.                                                             | O JSON enviado pelo cliente não está no formato esperado ou está faltando campos obrigatórios.                                                                                                           | "Invalid configuration format"                    |
| `SSH_CONNECTION_FAILED`          | Falha ao conectar ao servidor SSH.                                                                     | O host ou a porta estão inacessíveis, ou as credenciais fornecidas são inválidas para autenticação.                                                                                                       | "Failed to connect to SSH server"                |
| `SSH_SESSION_CREATION_FAILED`    | Falha ao criar uma nova sessão SSH.                                                                    | A conexão com o servidor SSH foi estabelecida, mas a criação da sessão falhou.                                                                                                                           | "Failed to create SSH session"                   |
| `STDIN_PIPE_FAILED`              | Falha ao abrir o pipe de entrada padrão.                                                               | Ocorre ao tentar abrir o stdin da sessão SSH para enviar dados do cliente.                                                                                                                               | "Failed to open stdin"                            |
| `STDOUT_PIPE_FAILED`             | Falha ao abrir o pipe de saída padrão.                                                                 | Ocorre ao tentar abrir o stdout da sessão SSH para ler dados enviados pelo servidor.                                                                                                                      | "Failed to open stdout"                           |
| `STDERR_PIPE_FAILED`             | Falha ao abrir o pipe de erro padrão.                                                                  | Ocorre ao tentar abrir o stderr da sessão SSH para capturar mensagens de erro do servidor.                                                                                                                | "Failed to open stderr"                           |
| `PTY_REQUEST_FAILED`             | Falha ao solicitar o Pseudo-Terminal (PTY).                                                            | A sessão SSH foi criada, mas não foi possível estabelecer o terminal interativo (PTY), geralmente devido a configurações do servidor remoto.                                                              | "Failed to request PTY"                           |
| `SHELL_START_FAILED`             | Falha ao iniciar o shell interativo.                                                                   | A conexão SSH e o PTY foram estabelecidos, mas o shell remoto não pôde ser iniciado.                                                                                                                      | "Failed to start shell"                           |
| `OUTPUT_READ_FAILED`             | Falha ao ler a saída da sessão SSH.                                                                    | Ocorre quando há um problema ao processar os dados de saída (stdout) enviados pelo servidor SSH, como desconexões inesperadas ou erros de leitura no buffer.                                               | "Failed to read SSH session output"              |

## Estrutura do Código

### `services`

- **StartSSHSession(ctx context.Context, configJSON string, conn *websocket.Conn) (*SSHSession, error)**: É responsável por estabelecer uma sessão SSH com a configuração especificada e a conexão WebSocket.
- **HandleOutput(ctx context.Context)**: É responsável por gerenciar a saída da sessão SSH e enviar os dados para a conexão WebSocket.
- **SendInput(input string)**: Envia a entrada especificada para a sessão SSH.
- **ResizeTerminal(rows, cols int)**: Redimensiona a janela do terminal da sessão SSH para o número especificado de linhas e colunas.
- **Close()**: Fecha a sessão SSH e libera quaisquer recursos associados.

### `handlers`

- **HandleWebSocket(ctx context.Context, w http.ResponseWriter, r \*http.Request)**: Gerencia a conexão WebSocket, fazendo o upgrade do protocolo HTTP para WebSocket, processando mensagens de configuração e comando, e estabelecendo uma sessão SSH.

### `clients`

- **GetClient(clientID string)**: Retorna o cliente associado ao `clientID`.
- **AddClient(clientID string, client \*Client)**: Adiciona um cliente ao mapa de clientes.
- **CleanupConnection(clientID string)**: Limpa a conexão do cliente, fechando a conexão SSH e WebSocket.

## Próximos Passos

1. **Melhorar tratativas de erros** :white_check_mark: :rocket:

   - Adicionar tratamento de erros para lidar com falhas na conexão SSH.

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

7. **Criar um GUI para SCP/RSYNC**:
   - Implementar uma interface gráfica para transferência de arquivos entre o cliente e o servidor SSH.

## Contribuições

Se você deseja contribuir com novas funcionalidades ou correções de bugs, fique à vontade para abrir uma "issue" ou enviar um "pull request". As sugestões acima são algumas das funcionalidades que podem ser adicionadas ao projeto, mas todas as contribuições são bem-vindas.

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Faça push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Este projeto está licenciado sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.
