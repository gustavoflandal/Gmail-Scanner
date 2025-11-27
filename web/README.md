# Gmail Scanner - Frontend

Interface web moderna para gerenciar e buscar e-mails do Gmail com tradução automática.

## Quick Start

### Desenvolvimento

```bash
# 1. Instalar dependências
npm install

# 2. Iniciar servidor de desenvolvimento
npm run dev
```

Acesse `http://localhost:5173`

O proxy automático redireciona requisições `/api/*` para `http://localhost:8080`

### Produção

```bash
# Build
npm run build

# Preview
npm run preview
```

Arquivos compilados em `dist/`

## Estrutura de Pastas

```
web/
├── src/
│   ├── components/          # Componentes reutilizáveis
│   │   ├── Header.jsx
│   │   ├── Footer.jsx
│   │   ├── LoadingSpinner.jsx
│   │   ├── Modal.jsx
│   │   └── Toast.jsx
│   ├── pages/              # Páginas (rotas)
│   │   ├── Home.jsx        # Landing page
│   │   ├── Dashboard.jsx   # Control center de varreduras
│   │   └── Messages.jsx    # Grid de e-mails
│   ├── services/           # Chamadas à API
│   │   └── api.js
│   ├── hooks/              # React Hooks customizados
│   │   └── useFetch.js
│   ├── App.jsx             # App principal com rotas
│   ├── main.jsx            # Entry point
│   └── index.css           # Estilos globais
├── public/                 # Assets estáticos
├── vite.config.js          # Config do Vite
├── tailwind.config.js      # Config do Tailwind
├── postcss.config.js       # Config do PostCSS
├── index.html              # HTML template
└── package.json            # Dependências
```

## Páginas

### Home (`/`)

- Landing page com features
- Status do sistema (API, Elasticsearch, Gmail)
- Como funciona
- Links para dashboard e autenticação

### Dashboard (`/dashboard`)

- Status da última varredura
- Data da próxima varredura agendada
- Botão para iniciar varredura manual
- Indicador de varredura em progresso
- Estatísticas (total de e-mails por pasta)
- Gráfico de distribuição de e-mails

### Messages (`/messages`)

- Grid com todos os e-mails indexados
- Paginação (20 por página)
- Barra de busca em tempo real
- Filtros por pasta
- Link para abrir e-mail no Gmail
- Ação para deletar e-mail com confirmação
- Exibe assunto original e traduzido

## Componentes

### Header
Navegação principal com:
- Logo do app
- Links para Dashboard e Mensagens
- Botão de Login

### Footer
Footer com:
- Informações do projeto
- Links úteis
- Créditos técnológicos

### LoadingSpinner
Indicador de carregamento com:
- Animação de rotação
- Mensagem customizável

### Modal
Diálogo reutilizável para:
- Confirmação de ações
- Delete warnings
- Inputs de usuário
- Botões customizáveis

### Toast
Notificações que:
- Auto-desaparecem após 3 segundos
- Suportam tipos: success, error, info, warning
- Botão manual de fechamento

## Serviços

### API Service (`src/services/api.js`)

```javascript
// Autenticação
apiService.login()                          // Redireciona para Google OAuth

// Saúde
apiService.health()                         // Status do sistema

// Varredura
apiService.startScan()                      // Inicia varredura manual
apiService.getScanStatus()                  // Status da varredura

// Mensagens
apiService.getMessages(page, query)         // Lista e-mails com busca
apiService.deleteMessage(messageId)         // Deleta um e-mail

// Estatísticas
apiService.getStats()                       // Stats do sistema
```

## Hooks

### useFetch

Hook customizado para fetch com gerenciamento automático de estado:

```javascript
const { data, loading, error, refetch } = useFetch(
  () => apiService.getMessages(1, ''),
  [page, search]
);
```

### useToast

Hook para gerenciar notificações:

```javascript
const { toasts, addToast, removeToast } = useToast();

addToast('Mensagem de sucesso!', 'success');
addToast('Erro!', 'error', 5000);
```

## Estilos

### Tailwind CSS

Usado para:
- Responsive design (mobile-first)
- Dark mode ready (configure em `tailwind.config.js`)
- Utility-first approach
- Custom colors em `tailwind.config.js`

Cores customizadas:
```css
primary-50, primary-100, primary-500, primary-600, primary-700
```

### CSS Global

Em `src/index.css`:
- Reset de estilos
- Scroll suave
- Tailwind directives

## Performance

### Otimizações Implementadas

1. **Code Splitting**: Vite faz code splitting automático
2. **Lazy Loading**: React Router permite lazy load de páginas
3. **Memoization**: Componentes podem ser memoizados se necessário
4. **Pagination**: Grid com paginação para não sobrecarregar
5. **API Caching**: Implementado em níveis (browser cache + axios)

### Build Otimizado

```bash
npm run build
# Saída:
# - dist/index.html
# - dist/assets/*.js (minified)
# - dist/assets/*.css (minified)
```

## Desenvolvimento

### Adicionar Nova Página

1. Criar arquivo em `src/pages/NovaPage.jsx`
2. Importar no `App.jsx`
3. Adicionar rota:

```jsx
<Route path="/nova" element={<NovaPage />} />
```

4. Atualizar Header com link

### Adicionar Novo Componente

1. Criar em `src/components/MeuComponente.jsx`
2. Importar e usar em páginas
3. Exportar: `export const MeuComponente = ({ props }) => { ... }`

### Adicionar Novo Hook

1. Criar em `src/hooks/useMeuHook.js`
2. Usar em componentes:

```javascript
const { algo } = useMeuHook();
```

## Debugging

### Console Browser

```javascript
// Ver chamadas à API
// DevTools > Network > XHR
```

### Vite Debug

```bash
# Verbose logging
npm run dev -- --debug
```

### React DevTools

Instale extensão no navegador para inspecionar componentes

## Variáveis de Ambiente

Criar `.env` na pasta `web/`:

```env
VITE_API_URL=http://localhost:8080
```

Usar no código:
```javascript
import.meta.env.VITE_API_URL
```

## Build e Deployment

### Build Local

```bash
npm run build
npm run preview
```

### Build em Docker

```bash
docker build -t gmail-scanner .
docker run -p 8080:8080 gmail-scanner
```

## Dependências

- **react** (18.2.0): Framework UI
- **react-dom** (18.2.0): Renderização DOM
- **react-router-dom** (6.20.0): Roteamento
- **axios** (1.6.2): HTTP client
- **tailwindcss** (3.3.6): Utility CSS
- **vite** (5.0.8): Build tool
- **@vitejs/plugin-react** (4.2.1): Plugin React para Vite

## Troubleshooting

### "Cannot GET /"

- Frontend não está servindo arquivos estáticos
- Certifique-se de que `npm run build` foi executado
- Verifique se backend está servindo arquivos de `web/public`

### CORS errors

- Backend precisa ter CORS configurado
- Verifique `corsMiddleware` em `cmd/api/main.go`
- Verifique proxy em `vite.config.js`

### API calls failing

- Certifique-se que backend está rodando em `http://localhost:8080`
- Verifique `/api/health` no navegador
- Ver logs do backend: `docker-compose logs backend`

### Hot reload não funciona

- Verifique se `npm run dev` está rodando
- Limpe cache do navegador (Ctrl+Shift+Del)
- Reinicie Vite: `npm run dev`

## Recursos Adicionais

- [Vite Documentation](https://vitejs.dev/)
- [React Documentation](https://react.dev/)
- [React Router](https://reactrouter.com/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Axios Documentation](https://axios-http.com/)

