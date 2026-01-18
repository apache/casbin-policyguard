import React, {Component} from "react";
import {Link, Redirect, Route, Switch, withRouter} from "react-router-dom";
import { Settings, LogOut, ChevronDown } from "lucide-react";
import "./App.css";
import * as Setting from "./Setting";
import * as AccountBackend from "./backend/AccountBackend";
import AuthCallback from "./AuthCallback";
import * as Conf from "./Conf";
import HomePage from "./HomePage";
import DatasetListPage from "./DatasetListPage";
import DatasetEditPage from "./DatasetEditPage";
import SigninPage from "./SigninPage";
import i18next from "i18next";
import SelectLanguageBox from "./SelectLanguageBox";
import { Avatar, AvatarFallback, AvatarImage } from "./components/ui/avatar";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "./components/ui/dropdown-menu";
import { Toaster } from "./components/ui/sonner";

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      selectedMenuKey: 0,
      account: undefined,
      uri: null,
    };

    Setting.initServerUrl();
    Setting.initCasdoorSdk(Conf.AuthConfig);
  }

  UNSAFE_componentWillMount() {
    this.updateMenuKey();
    this.getAccount();
  }

  componentDidUpdate() {
    // eslint-disable-next-line no-restricted-globals
    const uri = location.pathname;
    if (this.state.uri !== uri) {
      this.updateMenuKey();
    }
  }

  updateMenuKey() {
    // eslint-disable-next-line no-restricted-globals
    const uri = location.pathname;
    this.setState({
      uri: uri,
    });
    if (uri === "/") {
      this.setState({selectedMenuKey: "/"});
    } else if (uri.includes("/datasets")) {
      this.setState({selectedMenuKey: "/datasets"});
    } else {
      this.setState({selectedMenuKey: "null"});
    }
  }

  onUpdateAccount(account) {
    this.setState({
      account: account,
    });
  }

  setLanguage(account) {
    // let language = account?.language;
    const language = localStorage.getItem("language");
    if (language !== "" && language !== i18next.language) {
      Setting.setLanguage(language);
    }
  }

  getAccount() {
    AccountBackend.getAccount()
      .then((res) => {
        const account = res.data;
        if (account !== null) {
          this.setLanguage(account);
        }

        this.setState({
          account: account,
        });
      });
  }

  signout() {
    AccountBackend.signout()
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            account: null,
          });

          Setting.showMessage("success", "Successfully signed out, redirected to homepage");
          Setting.goToLink("/");
          // this.props.history.push("/");
        } else {
          Setting.showMessage("error", `Signout failed: ${res.msg}`);
        }
      });
  }

  renderAvatar() {
    if (this.state.account.avatar === "") {
      return (
        <Avatar>
          <AvatarFallback style={{backgroundColor: Setting.getAvatarColor(this.state.account.name)}}>
            {Setting.getShortName(this.state.account.name)}
          </AvatarFallback>
        </Avatar>
      );
    } else {
      return (
        <Avatar>
          <AvatarImage src={this.state.account.avatar} alt={Setting.getShortName(this.state.account.name)} />
          <AvatarFallback>
            {Setting.getShortName(this.state.account.name)}
          </AvatarFallback>
        </Avatar>
      );
    }
  }

  renderRightDropdown() {
    return (
      <DropdownMenu>
        <DropdownMenuTrigger className="flex items-center gap-2 cursor-pointer hover:bg-gray-100 px-3 py-2 rounded-md transition-colors">
          {this.renderAvatar()}
          {Setting.isMobile() ? null : <span>{Setting.getShortName(this.state.account.displayName)}</span>}
          <ChevronDown className="h-4 w-4" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => Setting.openLink(Setting.getMyProfileUrl(this.state.account))}>
            <Settings className="mr-2 h-4 w-4" />
            {i18next.t("account:My Account")}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => this.signout()}>
            <LogOut className="mr-2 h-4 w-4" />
            {i18next.t("account:Sign Out")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }

  renderAccount() {
    if (this.state.account === undefined) {
      return null;
    } else if (this.state.account === null) {
      return (
        <div className="flex items-center gap-4">
          <a href="/" className="hover:text-gray-600 transition-colors">
            {i18next.t("general:Home")}
          </a>
          <a href={Setting.getSigninUrl()} className="hover:text-gray-600 transition-colors">
            {i18next.t("account:Sign In")}
          </a>
          <a href={Setting.getSignupUrl()} className="hover:text-gray-600 transition-colors">
            {i18next.t("account:Sign Up")}
          </a>
        </div>
      );
    } else {
      return this.renderRightDropdown();
    }
  }

  renderMenu() {
    if (this.state.account === null || this.state.account === undefined) {
      return null;
    }

    return (
      <div className="flex items-center gap-6">
        <a
          href="/"
          className={`hover:text-gray-600 transition-colors ${
            this.state.selectedMenuKey === "/" ? "text-primary font-semibold" : ""
          }`}
        >
          {i18next.t("general:Home")}
        </a>
        <Link
          to="/datasets"
          className={`hover:text-gray-600 transition-colors ${
            this.state.selectedMenuKey === "/datasets" ? "text-primary font-semibold" : ""
          }`}
        >
          {i18next.t("general:Datasets")}
        </Link>
      </div>
    );
  }

  renderHomeIfSignedIn(component) {
    if (this.state.account !== null && this.state.account !== undefined) {
      return <Redirect to="/" />;
    } else {
      return component;
    }
  }

  renderSigninIfNotSignedIn(component) {
    if (this.state.account === null) {
      sessionStorage.setItem("from", window.location.pathname);
      return <Redirect to="/signin" />;
    } else if (this.state.account === undefined) {
      return null;
    } else {
      return component;
    }
  }

  renderContent() {
    return (
      <div>
        <header className="bg-white border-b border-gray-200 px-4 py-0 mb-1">
          <div className="flex items-center justify-between h-16">
            {
              Setting.isMobile() ? null : (
                <Link to={"/"}>
                  <div className="logo" />
                </Link>
              )
            }
            <div className="flex items-center gap-6 ml-auto">
              {this.renderMenu()}
              {this.renderAccount()}
              <SelectLanguageBox />
            </div>
          </div>
        </header>
        <Switch>
          <Route exact path="/callback" component={AuthCallback} />
          <Route exact path="/" render={(props) => <HomePage account={this.state.account} {...props} />} />
          <Route exact path="/signin" render={(props) => this.renderHomeIfSignedIn(<SigninPage {...props} />)} />
          <Route exact path="/datasets" render={(props) => this.renderSigninIfNotSignedIn(<DatasetListPage account={this.state.account} {...props} />)} />
          <Route exact path="/datasets/:datasetName" render={(props) => this.renderSigninIfNotSignedIn(<DatasetEditPage account={this.state.account} {...props} />)} />
        </Switch>
      </div>
    );
  }

  renderFooter() {
    // How to keep your footer where it belongs ?
    // https://www.freecodecamp.org/neyarnws/how-to-keep-your-footer-where-it-belongs-59c6aa05c59c/

    return (
      <footer id="footer" className="border-t border-gray-200 bg-white text-center py-4">
        Powered by <a className="font-bold text-black" target="_blank" rel="noreferrer" href="https://github.com/casbin/caswaf">CasWAF</a>
      </footer>
    );
  }

  render() {
    return (
      <div id="parent-area" className="relative min-h-screen">
        <Toaster />
        <div id="content-wrap" className="pb-[70px]">
          {
            this.renderContent()
          }
        </div>
        {
          this.renderFooter()
        }
      </div>
    );
  }
}

export default withRouter(App);
