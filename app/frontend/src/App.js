import React, { useState } from "react";
import "./App.css";
import {
  FormGroup,
  InputGroup,
  Button,
  Elevation,
  Callout,
  Intent,
  Classes,
  Card
} from "@blueprintjs/core";
import {
  HashRouter as Router,
  // BrowserRouter as Router,
  Switch,
  Route,
  Redirect,
  Link
} from "react-router-dom";

const SIGN_IN = "sign in";
const SIGN_IN_PATH = "/sign-in";
const REGISTER = "register";
const REGISTER_PATH = "/register";

function App() {
  return (
    <Router>
      <div className="App bp3-dark container" style={{ marginTop: 50 }}>
        <div className="row">
          <div className="col-xs-offset-2 col-xs-8 col-sm-offset-3 col-sm-6 col-lg-4 col-lg-offset-4">
            <h1 className="logo bp3-heading">
              <Link to="/">embly</Link>
            </h1>
            <div>
              <Switch>
                <Route path={SIGN_IN_PATH}>
                  <SignIn />
                </Route>
                <Route path={REGISTER_PATH}>
                  <Register />
                </Route>
                <Route path="/">
                  <Index />
                </Route>
              </Switch>
            </div>
          </div>
        </div>
      </div>
    </Router>
  );
}

async function getUser(setUser, setLoading) {
  const resp = await fetch("/api/auth/user", {
    method: "GET",
    headers: {
      "Content-Type": "application/json"
    }
  });
  const body = await resp.json();
  if (resp.status < 300) {
    setUser(body);
  } else {
    setUser(null);
  }
  setLoading(false);
}

async function signOut(setLoading, callback) {
  const resp = await fetch("/api/auth/sign-out", {
    method: "GET",
    headers: {
      "Content-Type": "application/json"
    }
  });
  const body = await resp.json();
  setLoading(false);
  callback();
}

class Index extends React.Component {
  constructor(props) {
    super(props);
    this.getUser = this.getUser.bind(this);
    this.state = {
      user: null,
      loading: true
    };
  }
  componentDidMount() {
    this.getUser();
  }
  getUser() {
    getUser(
      user => {
        this.setState({ user });
      },
      loading => {
        this.setState({ loading });
      }
    )
      .then()
      .catch(err => {
        console.log(err);
      });
  }
  render() {
    let skeletonClass = Classes.SKELETON;
    if (!this.state.loading) {
      skeletonClass = "";
    }
    if (this.state.user === null || this.state.loading === true) {
      return (
        <div>
          <p className={skeletonClass}>welcome to embly</p>
          <div>
            <ButtonLink
              to={SIGN_IN_PATH}
              intent={Intent.PRIMARY}
              className={skeletonClass}
            >
              {SIGN_IN}
            </ButtonLink>
            <ButtonLink
              to={REGISTER_PATH}
              intent={Intent.PRIMARY}
              className={skeletonClass}
            >
              {REGISTER}
            </ButtonLink>
          </div>
        </div>
      );
    }
    return (
      <div>
        <p className={skeletonClass}>welcome to embly</p>
        <div>
          You are logged with username "{this.state.user.username}" and email
          address "{this.state.user.email}"
        </div>
        <br />
        <div
          className={"bp3-button bp3-intent-" + Intent.PRIMARY}
          onClick={() => {
            signOut(loading => {
              this.setState({ loading });
            }, this.getUser);
          }}
        >
          Sign Out
        </div>
      </div>
    );
  }
}

const ButtonLink = ({ children, to, intent, className }) => {
  return (
    <Link
      className={className + " bp3-button bp3-intent-" + intent}
      style={{ marginRight: 10 }}
      to={to}
    >
      {children}
    </Link>
  );
};

async function signInRequest({ setResponse, username, password, setLoading }) {
  setLoading(true);
  try {
    const resp = await fetch("/api/auth/sign-in", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "include",
      body: JSON.stringify({ username, password })
    });
    const body = await resp.json();
    setLoading(false);
    setResponse(body);
  } catch (err) {
    setResponse({ err: err });
  }
  setLoading(false);
}

const renderCallout = response => {
  if (response && response.err) {
    return <Callout intent={Intent.WARNING}>{response.err.message}</Callout>;
  }
};

const SignIn = ({}) => {
  const [response, setResponse] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  if (response.success) {
    return <Redirect to="/" />;
  }
  const onSubmit = () => {
    signInRequest({
      setResponse,
      username,
      password,
      setLoading
    });
  };
  return (
    <Card interactive={true} elevation={Elevation.TWO}>
      <h2 style={{ marginTop: 0 }}>{SIGN_IN}</h2>
      {renderCallout(response)}
      <form
        onSubmit={e => {
          e.stopPropagation();
          onSubmit();
        }}
      >
        <AuthField
          response={response}
          label="username/email"
          name="username"
          setter={setUsername}
          value={username}
        />
        <AuthField
          response={response}
          type="password"
          name="password"
          setter={setPassword}
          value={password}
        />
        <p>
          <Button
            type="submit"
            onClick={onSubmit}
            loading={loading}
            intent={Intent.PRIMARY}
          >
            Submit
          </Button>
        </p>
        <p>
          <Link to={REGISTER_PATH}>{REGISTER}</Link>
        </p>
      </form>
    </Card>
  );
};

async function registerRequest({
  setResponse,
  username,
  email,
  password,
  setLoading
}) {
  setLoading(true);
  const resp = await fetch("/api/auth/register", {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ username, email, password })
  });
  console.log(resp);
  const body = await resp.json();
  setLoading(false);
  setResponse(body);
}

const msgForResponse = (field, response) => {
  if (response.field === field) {
    return response.msg;
  }
  return "";
};

const AuthField = ({
  response,
  name,
  label,
  type,
  helperText,
  setter,
  value
}) => {
  return (
    <FormGroup
      helperText={msgForResponse(name, response) || helperText}
      label={label || name}
      labelFor={name}
      labelInfo=""
      intent={msgForResponse(name, response) ? Intent.WARNING : Intent.NONE}
    >
      <InputGroup
        onChange={e => setter(e.target.value)}
        id={name}
        placeholder=""
        type={type}
        intent={msgForResponse(name, response) ? Intent.WARNING : Intent.NONE}
        value={value}
      />
    </FormGroup>
  );
};

const Register = ({}) => {
  const [response, setResponse] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  if (response.success) {
    return <Redirect to="/" />;
  }
  const onSubmit = () => {
    registerRequest({
      setResponse,
      username,
      email,
      password,
      setLoading
    });
  };
  return (
    <Card interactive={true} elevation={Elevation.TWO}>
      <h2 style={{ marginTop: 0 }}>{REGISTER}</h2>
      <form
        onSubmit={e => {
          console.log("hmmm");
          e.stopPropagation();
          onSubmit();
        }}
      >
        <AuthField
          response={response}
          name="username"
          helperText="must be at least 3 characters, can't contain most special characters"
          setter={setUsername}
          value={username}
        />
        <AuthField
          response={response}
          name="email"
          type="email"
          label="email address"
          setter={setEmail}
          value={email}
        />
        <AuthField
          response={response}
          name="password"
          helperText="must be at least 8 characters long, palindromes are not allowed"
          setter={setPassword}
          value={password}
          type="password"
        />
        <p>
          <Button
            type="submit"
            onClick={onSubmit}
            loading={loading}
            intent={Intent.PRIMARY}
          >
            Submit
          </Button>
        </p>
      </form>
      <p>
        <Link to={SIGN_IN_PATH}>{SIGN_IN}</Link>
      </p>
    </Card>
  );
};

export default App;
