import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';
import { useToast } from '../components/Toast';

export const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [showInstructions, setShowInstructions] = useState(false);
  const { toasts, addToast } = useToast();
  const navigate = useNavigate();

  const handleLogin = async (e) => {
    e.preventDefault();
    
    if (!email || !password) {
      addToast('Por favor, preencha email e senha', 'error');
      return;
    }

    setLoading(true);
    try {
      await apiService.loginWithIMAP(email, password);
      addToast('Login realizado com sucesso!', 'success');
      setTimeout(() => navigate('/dashboard'), 1000);
    } catch (error) {
      addToast(error.response?.data?.error || 'Erro ao fazer login. Verifique suas credenciais.', 'error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[80vh] flex items-center justify-center px-4">
      {/* Toasts */}
      <div className="fixed top-4 right-4 space-y-2 z-50">
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={`${
              toast.type === 'success'
                ? 'bg-green-500'
                : toast.type === 'error'
                ? 'bg-red-500'
                : 'bg-blue-500'
            } text-white px-4 py-3 rounded-lg shadow-lg`}
          >
            {toast.message}
          </div>
        ))}
      </div>

      <div className="max-w-md w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="text-6xl mb-4">📧</div>
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Gmail Scanner</h1>
          <p className="text-gray-600">
            Acesse seus e-mails com IMAP
          </p>
        </div>

        {/* Login Form */}
        <div className="bg-white rounded-lg shadow-lg p-8">
          <form onSubmit={handleLogin} className="space-y-6">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                Email do Gmail
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="seu.email@gmail.com"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-600"
                required
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-2">
                Senha de App
                <button
                  type="button"
                  onClick={() => setShowInstructions(!showInstructions)}
                  className="ml-2 text-primary-600 hover:text-primary-700 text-xs"
                >
                  {showInstructions ? '(ocultar ajuda)' : '(como obter?)'}
                </button>
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="xxxx xxxx xxxx xxxx"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-600"
                required
              />
            </div>

            {showInstructions && (
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-sm">
                <h4 className="font-semibold text-blue-900 mb-2">Como gerar senha de app:</h4>
                <ol className="list-decimal list-inside space-y-1 text-blue-800">
                  <li>Acesse: <a href="https://myaccount.google.com/apppasswords" target="_blank" rel="noopener noreferrer" className="underline">myaccount.google.com/apppasswords</a></li>
                  <li>Ative a verificação em duas etapas (se ainda não tiver)</li>
                  <li>Selecione "Email" como app e "Outro" como dispositivo</li>
                  <li>Dê um nome (ex: "Gmail Scanner")</li>
                  <li>Copie a senha de 16 caracteres gerada</li>
                  <li>Cole aqui sem espaços</li>
                </ol>
                <p className="mt-2 text-xs text-blue-700">
                  <strong>Importante:</strong> Você também precisa habilitar IMAP nas configurações do Gmail
                  (Configurações → Encaminhamento e POP/IMAP → Ativar IMAP)
                </p>
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-primary-600 text-white px-6 py-3 rounded-lg font-semibold hover:bg-primary-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition"
            >
              {loading ? 'Conectando...' : 'Entrar'}
            </button>
          </form>
        </div>

        {/* Info Cards */}
        <div className="mt-8 grid grid-cols-1 gap-4">
          <div className="bg-green-50 border border-green-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <span className="text-2xl">✅</span>
              <div>
                <h3 className="font-semibold text-green-900">100% Seguro</h3>
                <p className="text-sm text-green-800">
                  Usamos protocolo IMAP seguro (SSL/TLS). Suas credenciais são criptografadas.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <span className="text-2xl">🔒</span>
              <div>
                <h3 className="font-semibold text-blue-900">Senha de App</h3>
                <p className="text-sm text-blue-800">
                  Use uma senha de app do Google, não sua senha principal. Mais seguro e específico para este app.
                </p>
              </div>
            </div>
          </div>

          <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <span className="text-2xl">⚡</span>
              <div>
                <h3 className="font-semibold text-purple-900">Sem OAuth</h3>
                <p className="text-sm text-purple-800">
                  Sem necessidade de aprovação do Google. Configure em 2 minutos e comece a usar!
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
