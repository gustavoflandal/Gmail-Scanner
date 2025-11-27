export const Footer = () => {
  return (
    <footer className="bg-gray-800 text-gray-300 mt-12">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div>
            <h3 className="text-white font-bold mb-4">Gmail Scanner</h3>
            <p className="text-sm">Aplicação de código aberto para varrer e organizar seus e-mails do Gmail.</p>
          </div>

          <div>
            <h4 className="text-white font-bold mb-4">Links Rápidos</h4>
            <ul className="text-sm space-y-2">
              <li><a href="#" className="hover:text-white transition">Documentação</a></li>
              <li><a href="#" className="hover:text-white transition">GitHub</a></li>
              <li><a href="#" className="hover:text-white transition">Issues</a></li>
            </ul>
          </div>

          <div>
            <h4 className="text-white font-bold mb-4">Tecnologias</h4>
            <p className="text-sm">
              100% gratuito e open source. Desenvolvido com Go, React, OpenSearch e LibreTranslate.
            </p>
          </div>
        </div>

        <div className="border-t border-gray-700 mt-8 pt-8 text-center text-sm">
          <p>&copy; 2024 Gmail Scanner. Todos os direitos reservados.</p>
        </div>
      </div>
    </footer>
  );
};
