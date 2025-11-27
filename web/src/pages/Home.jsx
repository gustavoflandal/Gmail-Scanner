import { Link } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { apiService } from '../services/api';

export const Home = () => {
  const [health, setHealth] = useState(null);
  const [authError, setAuthError] = useState(null);

  useEffect(() => {
    // Verificar se há erro de autenticação na URL
    const params = new URLSearchParams(window.location.search);
    const error = params.get('auth_error');
    if (error) {
      const errorMessages = {
        'access_denied': 'Você negou o acesso. Por favor, tente novamente.',
        'exchange_failed': 'Falha ao trocar código por token. Tente novamente.',
        'user_info_failed': 'Falha ao obter informações do usuário. Tente novamente.',
        'invalid_state': 'Sessão inválida. Por favor, tente novamente.',
        'missing_params': 'Parâmetros faltando. Por favor, tente novamente.',
        'jwt_failed': 'Falha ao gerar token de autenticação. Tente novamente.',
      };
      setAuthError(errorMessages[error] || `Erro de autenticação: ${error}`);
      // Limpar parâmetro da URL
      window.history.replaceState({}, document.title, window.location.pathname);
    }

    const checkHealth = async () => {
      try {
        const data = await apiService.health();
        setHealth(data);
      } catch (error) {
        console.error('Health check failed:', error);
      }
    };
    checkHealth();
  }, []);

  return (
    <div className="max-w-7xl mx-auto px-4 py-16">
      {/* Auth Error Alert */}
      {authError && (
        <div className="mb-8 bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
          <span className="text-xl">⚠️</span>
          <div>
            <h3 className="font-semibold text-red-800">Erro de Autenticação</h3>
            <p className="text-red-700">{authError}</p>
          </div>
        </div>
      )}

      {/* Hero Section */}
      <div className="text-center mb-16">
        <div className="text-6xl mb-4">📧</div>
        <h1 className="text-5xl font-bold text-gray-900 mb-4">Gmail Scanner</h1>
        <p className="text-xl text-gray-600 max-w-2xl mx-auto mb-8">
          Organize e encontre seus e-mails do Gmail com facilidade. Varredura automática, tradução inteligente e busca poderosa.
        </p>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link
            to="/login"
            className="bg-primary-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-primary-700 transition"
          >
            Fazer Login
          </Link>
          <Link
            to="/dashboard"
            className="bg-white text-primary-600 border-2 border-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-primary-50 transition"
          >
            Ver Dashboard
          </Link>
        </div>
      </div>

      {/* Features */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-16">
        <div className="bg-white p-6 rounded-lg shadow-lg">
          <div className="text-4xl mb-4">🔍</div>
          <h3 className="text-xl font-bold text-gray-900 mb-2">Busca Poderosa</h3>
          <p className="text-gray-600">
            Encontre e-mails instantaneamente com filtros avançados por pasta, data e conteúdo.
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-lg">
          <div className="text-4xl mb-4">🤖</div>
          <h3 className="text-xl font-bold text-gray-900 mb-2">Tradução Automática</h3>
          <p className="text-gray-600">
            Todos os assuntos são automaticamente traduzidos para português brasileiro.
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-lg">
          <div className="text-4xl mb-4">⚙️</div>
          <h3 className="text-xl font-bold text-gray-900 mb-2">Varredura Automática</h3>
          <p className="text-gray-600">
            Execute varreduras automáticas a cada 6 horas ou manualmente sempre que desejar.
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-lg">
          <div className="text-4xl mb-4">🔐</div>
          <h3 className="text-xl font-bold text-gray-900 mb-2">Seguro e Privado</h3>
          <p className="text-gray-600">
            Autenticação via OAuth 2.0. Seus dados permanecem seguros e privados.
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-lg">
          <div className="text-4xl mb-4">💾</div>
          <h3 className="text-xl font-bold text-gray-900 mb-2">Armazenamento Local</h3>
          <p className="text-gray-600">
            Use OpenSearch para armazenar e indexar seus e-mails localmente.
          </p>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-lg">
          <div className="text-4xl mb-4">🆓</div>
          <h3 className="text-xl font-bold text-gray-900 mb-2">100% Gratuito</h3>
          <p className="text-gray-600">
            Código aberto e sem custos. Desenvolvido com tecnologias open source.
          </p>
        </div>
      </div>

      {/* Status */}
      {health && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Status do Sistema</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <p className="text-sm text-gray-600 mb-2">API</p>
              <p className="text-lg font-semibold flex items-center gap-2">
                <span className={`h-3 w-3 rounded-full ${health.status === 'ok' ? 'bg-green-500' : 'bg-red-500'}`}></span>
                {health.status === 'ok' ? 'Online' : 'Offline'}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-600 mb-2">Elasticsearch</p>
              <p className="text-lg font-semibold flex items-center gap-2">
                <span className={`h-3 w-3 rounded-full ${health.services?.elasticsearch ? 'bg-green-500' : 'bg-red-500'}`}></span>
                {health.services?.elasticsearch ? 'Conectado' : 'Desconectado'}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-600 mb-2">Gmail</p>
              <p className="text-lg font-semibold flex items-center gap-2">
                <span className={`h-3 w-3 rounded-full ${health.services?.gmail ? 'bg-green-500' : 'bg-orange-500'}`}></span>
                {health.services?.gmail ? 'Autenticado' : 'Não autenticado'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* How it works */}
      <div className="mt-16">
        <h2 className="text-3xl font-bold text-gray-900 mb-8 text-center">Como Funciona</h2>
        <div className="bg-white rounded-lg shadow-lg p-8">
          <ol className="space-y-6">
            <li className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
                1
              </div>
              <div>
                <h3 className="font-bold text-lg text-gray-900">Autentique-se</h3>
                <p className="text-gray-600">Faça login de forma segura usando sua conta Google via OAuth 2.0.</p>
              </div>
            </li>
            <li className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
                2
              </div>
              <div>
                <h3 className="font-bold text-lg text-gray-900">Iniciare Varredura</h3>
                <p className="text-gray-600">Inicie uma varredura manual ou deixe o sistema fazer automaticamente a cada 6 horas.</p>
              </div>
            </li>
            <li className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
                3
              </div>
              <div>
                <h3 className="font-bold text-lg text-gray-900">Tradução Automática</h3>
                <p className="text-gray-600">Os assuntos dos e-mails são automaticamente traduzidos para português.</p>
              </div>
            </li>
            <li className="flex gap-4">
              <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
                4
              </div>
              <div>
                <h3 className="font-bold text-lg text-gray-900">Busque e Organize</h3>
                <p className="text-gray-600">Use a busca poderosa para encontrar e-mails rapidamente e clique para abrir no Gmail.</p>
              </div>
            </li>
          </ol>
        </div>
      </div>
    </div>
  );
};
