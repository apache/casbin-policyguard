import React from "react";
import { Button } from "./components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./components/ui/card";
import {withRouter} from "react-router-dom";
import * as Setting from "./Setting";

class AuthCallback extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      msg: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.login();
  }

  getFromLink() {
    const from = sessionStorage.getItem("from");
    if (from === null) {
      return "/";
    }
    return from;
  }

  login() {
    Setting.signin().then((res) => {
      if (res.status === "ok") {
        Setting.showMessage("success", "Logged in successfully");

        const link = this.getFromLink();
        Setting.goToLink(link);
      } else {
        this.setState({
          msg: res.msg,
        });
      }
    });
  }

  render() {
    return (
      <div className="flex justify-center items-center min-h-screen">
        {this.state.msg === null ? (
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-lg">Signing in...</p>
          </div>
        ) : (
          <Card className="w-96">
            <CardHeader>
              <CardTitle className="text-red-600">Login Error</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="mb-4">{this.state.msg}</p>
              <div className="flex gap-2">
                <Button>Details</Button>
                <Button variant="outline">Help</Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    );
  }
}

export default withRouter(AuthCallback);
