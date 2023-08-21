import './App.css';
import { Router, Route, Routes } from '@solidjs/router';
import Home from './pages/home/Home';
import Room from './pages/room/Room';

const App = () => {
  return (
    <div class="App">
      <Router>
        <Routes>
          <Route path="/" component={Home} />
          <Route path="/room/:code" component={Room} />
        </Routes>
      </Router>
    </div>
  );
};

export default App;
