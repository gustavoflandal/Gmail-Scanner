import { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { Modal } from '../components/Modal';
import { useToast } from '../components/Toast';

export const Messages = () => {
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [folderFilter, setFolderFilter] = useState('');
  const [folders, setFolders] = useState([]);
  const [totalPages, setTotalPages] = useState(1);
  const [totalEmails, setTotalEmails] = useState(0);
  const [deleteModal, setDeleteModal] = useState({ isOpen: false, messageId: null });
  const { toasts, addToast } = useToast();

  const fetchMessages = async (pageNum = 1, query = '') => {
    try {
      setLoading(true);
      const data = await apiService.getMessages(pageNum, query);
      setMessages(data.emails || []);
      setTotalPages(data.total_pages || 1);
      setTotalEmails(data.total || 0);
      
      // Extrair pastas únicas
      const uniqueFolders = [...new Set((data.emails || []).map(m => m.folder))];
      setFolders(uniqueFolders);
    } catch (error) {
      addToast('Erro ao carregar mensagens: ' + error.message, 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMessages(page, search);
  }, [page, search]);

  const handleSearch = (e) => {
    const query = e.target.value;
    setSearch(query);
    setPage(1);
  };

  // Filtrar mensagens por pasta localmente
  const filteredMessages = folderFilter
    ? messages.filter(msg => msg.folder === folderFilter)
    : messages;

  const handleDelete = async () => {
    try {
      await apiService.deleteMessage(deleteModal.messageId);
      addToast('E-mail deletado com sucesso!', 'success');
      setDeleteModal({ isOpen: false, messageId: null });
      fetchMessages(page, search);
    } catch (error) {
      addToast('Erro ao deletar e-mail: ' + error.message, 'error');
    }
  };

  const formatDate = (dateString) => {
    try {
      return new Date(dateString).toLocaleDateString('pt-BR', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return 'Inválido';
    }
  };

  if (loading && messages.length === 0) {
    return <LoadingSpinner message="Carregando mensagens..." />;
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Toasts */}
      <div className="fixed top-4 right-4 space-y-2 z-40">
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

      {/* Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Mensagens</h1>
          <p className="text-gray-600">{totalEmails} e-mails indexados</p>
        </div>
        <button
          onClick={() => fetchMessages(page, search)}
          className="bg-primary-600 text-white px-4 py-2 rounded-lg hover:bg-primary-700 transition"
        >
          🔄 Atualizar
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-lg p-6 mb-8 space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Buscar
          </label>
          <input
            type="text"
            placeholder="Buscar por assunto, remetente ou conteúdo..."
            value={search}
            onChange={handleSearch}
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-600"
          />
        </div>
        
        {folders.length > 0 && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Filtrar por Pasta
            </label>
            <select
              value={folderFilter}
              onChange={(e) => setFolderFilter(e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-600"
            >
              <option value="">Todas as pastas</option>
              {folders.map((folder) => (
                <option key={folder} value={folder}>
                  {folder}
                </option>
              ))}
            </select>
          </div>
        )}
        
        {(search || folderFilter) && (
          <div className="flex gap-2">
            {search && (
              <span className="bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm">
                Busca: {search}
                <button
                  onClick={() => setSearch('')}
                  className="ml-2 text-blue-600 hover:text-blue-800"
                >
                  ✕
                </button>
              </span>
            )}
            {folderFilter && (
              <span className="bg-green-100 text-green-800 px-3 py-1 rounded-full text-sm">
                Pasta: {folderFilter}
                <button
                  onClick={() => setFolderFilter('')}
                  className="ml-2 text-green-600 hover:text-green-800"
                >
                  ✕
                </button>
              </span>
            )}
          </div>
        )}
      </div>

      {/* Messages Table */}
      <div className="bg-white rounded-lg shadow-lg overflow-hidden">
        {filteredMessages.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            <p>Nenhuma mensagem encontrada.</p>
            {search && <p className="text-sm mt-2">Tente refinar sua busca.</p>}
            {folderFilter && <p className="text-sm mt-2">Nenhum email nesta pasta.</p>}
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-100 border-b">
                  <tr>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">Data</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">Pasta</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">De</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">Assunto</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">Preview</th>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-gray-900">Ações</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredMessages.map((msg) => (
                    <tr key={msg.message_id} className="border-b hover:bg-gray-50 transition">
                      <td className="px-6 py-4 text-sm text-gray-600 whitespace-nowrap">
                        {formatDate(msg.timestamp)}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs font-medium whitespace-nowrap">
                          {msg.folder}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-600 max-w-[200px]">
                        <div className="truncate" title={msg.from}>
                          {msg.from}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm font-medium text-gray-900 max-w-[300px]">
                        <div className="truncate" title={msg.subject}>
                          {msg.subject}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500 max-w-[250px]">
                        <div className="truncate" title={msg.snippet_preview}>
                          {msg.snippet_preview}
                        </div>
                      </td>
                      <td className="px-6 py-4 text-sm whitespace-nowrap">
                        <div className="flex gap-2">
                          <a
                            href={msg.link}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-primary-600 hover:text-primary-700 font-medium"
                          >
                            Abrir
                          </a>
                          <button
                            onClick={() => setDeleteModal({ isOpen: true, messageId: msg.message_id })}
                            className="text-red-600 hover:text-red-700 font-medium"
                          >
                            Deletar
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            <div className="px-6 py-4 bg-gray-50 border-t flex items-center justify-between">
              <span className="text-sm text-gray-600">
                Mostrando {filteredMessages.length} de {totalEmails} emails | Página {page} de {totalPages}
              </span>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 disabled:bg-gray-100 disabled:cursor-not-allowed transition"
                >
                  ← Anterior
                </button>
                <button
                  onClick={() => setPage((p) => (p < totalPages ? p + 1 : p))}
                  disabled={page === totalPages}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 disabled:bg-gray-100 disabled:cursor-not-allowed transition"
                >
                  Próxima →
                </button>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Delete Modal */}
      <Modal
        isOpen={deleteModal.isOpen}
        title="Deletar E-mail"
        onClose={() => setDeleteModal({ isOpen: false, messageId: null })}
        onConfirm={handleDelete}
        confirmText="Deletar"
        cancelText="Cancelar"
        isDangerous
      >
        <p>Tem certeza que deseja deletar este e-mail? Esta ação não pode ser desfeita.</p>
      </Modal>
    </div>
  );
};
