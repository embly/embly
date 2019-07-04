import React, { Component } from "react";
import "./App.css";
import { Callout, Card, Elevation, H5 } from "@blueprintjs/core";

class App extends Component {
  render() {
    return (
      <div className="container">
        <H5>Embly</H5>
        <Card interactive={false} elevation={Elevation.TWO}>
          <Callout title={"Hello"}>words</Callout>
        </Card>
      </div>
    );
  }
}

export default App;
