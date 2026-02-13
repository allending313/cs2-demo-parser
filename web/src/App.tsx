import { Route, Switch } from "wouter";
import DemoUpload from "./components/DemoUpload";
import MatchPage from "./components/MatchPage";
import "./index.css";

export default function App() {
  return (
    <Switch>
      <Route path="/match/:id" component={MatchPage} />
      <Route path="/" component={DemoUpload} />
    </Switch>
  );
}
