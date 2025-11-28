import { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Header } from './components/Header';
import { Footer } from './components/Footer';
import { Home } from './pages/Home';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import Articles from './pages/Articles';
import ReadArticle from './pages/ReadArticle';
import { startStorageMonitoring, cleanupStorage } from './utils/storage';

function App() {
  useEffect(() => {
    // Limpar storage na inicialização
    cleanupStorage();

    // Monitorar uso de storage a cada 60 segundos
    startStorageMonitoring(60000);
  }, []);

  return (
    <Router>
      <div className="flex flex-col min-h-screen">
        <Header />
        <main className="flex-1">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/articles" element={<Articles />} />
            <Route path="/read/:id" element={<ReadArticle />} />
          </Routes>
        </main>
        <Footer />
      </div>
    </Router>
  );
}

export default App;
