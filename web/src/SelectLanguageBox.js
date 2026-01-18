import React from "react";
import * as Setting from "./Setting";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "./components/ui/dropdown-menu";
import { Globe } from "lucide-react";

const LanguageItems = [
  {lang: "en", label: "English"},
  {lang: "zh", label: "中文"},
];

class SelectLanguageBox extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  render() {
    return (
      <DropdownMenu>
        <DropdownMenuTrigger className="flex items-center justify-center w-11 h-11 hover:bg-gray-100 rounded-md transition-colors cursor-pointer">
          <Globe className="h-5 w-5" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {LanguageItems.map(({lang, label}) => (
            <DropdownMenuItem key={lang} onClick={() => Setting.changeLanguage(lang)}>
              {label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }
}

export default SelectLanguageBox;
