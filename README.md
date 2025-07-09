# Blaze WebSocket Client - Go

Este é um cliente WebSocket para a plataforma Blaze, portado do TypeScript para Go.

## Funcionalidades Principais

### 1. Conexão em Tempo Real
- Conecta aos jogos da Blaze (Crash, Double, Chat)
- Suporte a diferentes tipos de crash (crash, crash_2, crash_neymarjr)
- Reconexão automática opcional

### 2. 🆕 GetNextGameEventTick - Aguardar Jogo Completo
Nova função que aguarda um jogo completo e retorna todos os eventos:

**Como funciona:**
1. Conecta ao jogo especificado
2. Aguarda o próximo jogo começar (status `waiting`)
3. Coleta todos os eventos `crash.tick` dessa rodada
4. Quando o jogo termina (status `complete`), fecha a conexão
5. Retorna todos os eventos coletados

**Com timeout:**
```go
// Timeout de 30 segundos
ctx, cancel := context.WithTimeout(context.Background(), 160*time.Second)
defer cancel()

result, err := GetNextGameEventTickWithContext(ctx, "crash")
```
## Tipos de Conexão

### Web Types
- `"blaze"` - Para jogos (crash, double, etc.)
- `"blaze-chat"` - Para chat

### Game Types (apenas para web "blaze")
- `"crash"` - Crash original
- `"doubles"` - Double
- `"crash_2"` - Crash 2 (com apostas e bonus rounds)
- `"crash_neymarjr"` - Crash Neymar Jr

## Opções de Conexão

```go
conn, err := MakeConnection(Connection{
    GameType: "crash",
    Web:      "blaze",
    Token:    &token,                    // Token de autenticação (opcional)
    URL:      &customURL,                // URL customizada (opcional)
    TimeoutPing: &timeout,               // Timeout do ping em ms (padrão: 10000)
    CacheIgnoreRepeatedEvents: &false,   // Desabilitar cache (padrão: true)
    Options: &ConnectionOptions{
        Host:    &host,                  // Host customizado
        Origin:  &origin,                // Origin customizado
        Headers: map[string]string{      // Headers customizados
            "Custom-Header": "value",
        },
    },
})
```

## Eventos Disponíveis

### Crash
- `crash.tick` - Tick do crash com informações do jogo
- `crash.tick-bets` - Apostas do crash (apenas crash_2)

### Double
- `double.tick` - Tick do double com informações do jogo

### Chat
- `chat.message` - Mensagens do chat

### Sistema
- `subscriptions` - Lista de subscrições ativas
- `close` - Evento de fechamento da conexão

## Estrutura de Dados

### CrashTickEvent
```go
type CrashTickEvent struct {
    ID           string   `json:"id"`
    UpdatedAt    string   `json:"updated_at"`
    Status       string   `json:"status"`        // "waiting", "playing", "complete"
    CrashPoint   *float64 `json:"crash_point"`   // Multiplicador (null durante jogo)
    IsBonusRound bool     `json:"is_bonus_round"` // Apenas crash_2
}
```

### GameEventResult
```go
type GameEventResult struct {
    Events []CrashTickEvent `json:"events"`
    Error  error           `json:"error,omitempty"`
}
```

## Testando

```bash
# Executar todos os testes
go test -v

# Testar apenas timeout (rápido)
go test -v -run TestGetNextGameEventTickTimeout

# Testar conexão básica
go test -v -run TestConnection
```

## Executando

```bash
# Compilar
go build .

# Executar exemplo principal
./blaze

# Executar demo
# (altere a função main para chamar DemoGetNextGameEventTick)
```

## Vantagens da Nova Função GetNextGameEventTick

1. **Simplicidade**: Uma única chamada para obter um jogo completo
2. **Gerenciamento Automático**: Conecta, coleta eventos e desconecta automaticamente
3. **Controle de Timeout**: Suporte nativo a contexto para timeout
4. **Dados Estruturados**: Retorna eventos tipados prontos para uso
5. **Error Handling**: Tratamento robusto de erros e timeouts

## Casos de Uso Ideais

- **Análise de Jogos**: Coletar dados de jogos completos para análise
- **Monitoramento**: Acompanhar resultados de múltiplos jogos
- **Alertas**: Detectar padrões específicos em jogos
- **Logging**: Registrar histórico de jogos automaticamente

## Arquitetura

O projeto segue a mesma arquitetura limpa do TypeScript:
- **Domain**: Interfaces e contratos
- **Data**: Implementações dos casos de uso
- **Infra**: Implementações de infraestrutura (WebSocket, URLs)

## Dependências

- `github.com/gorilla/websocket` - Cliente WebSocket para Go
