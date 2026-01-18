import React from "react";
import { Button } from "./components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./components/ui/card";
import { Input } from "./components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./components/ui/select";
import { DatePicker } from "./components/ui/date-picker";
import * as DatasetBackend from "./backend/DatasetBackend";
import * as Setting from "./Setting";
import { parse, format } from "date-fns";
import i18next from "i18next";

class DatasetEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      datasetName: props.match.params.datasetName,
      dataset: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getDataset();
  }

  getDataset() {
    DatasetBackend.getDataset(this.props.account.name, this.state.datasetName)
      .then((dataset) => {
        this.setState({
          dataset: dataset,
        });
      });
  }

  parseDatasetField(key, value) {
    if (["score"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateDatasetField(key, value) {
    value = this.parseDatasetField(key, value);

    const dataset = this.state.dataset;
    dataset[key] = value;
    this.setState({
      dataset: dataset,
    });
  }

  renderDataset() {
    const startDate = this.state.dataset.startDate ? parse(this.state.dataset.startDate, "yyyy-MM-dd", new Date()) : null;
    const endDate = this.state.dataset.endDate ? parse(this.state.dataset.endDate, "yyyy-MM-dd", new Date()) : null;

    return (
      <Card className="ml-1">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>{i18next.t("dataset:Edit Dataset")}</CardTitle>
            <Button onClick={this.submitDatasetEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("general:Name")}:</label>
            <div className="col-span-10">
              <Input
                value={this.state.dataset.name}
                onChange={e => this.updateDatasetField("name", e.target.value)}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Start date")}:</label>
            <div className="col-span-4">
              <DatePicker
                date={startDate}
                onDateChange={(date) => this.updateDatasetField("startDate", date ? format(date, "yyyy-MM-dd") : "")}
              />
            </div>
            <label className="col-span-2">{i18next.t("dataset:End date")}:</label>
            <div className="col-span-4">
              <DatePicker
                date={endDate}
                onDateChange={(date) => this.updateDatasetField("endDate", date ? format(date, "yyyy-MM-dd") : "")}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Full name")}:</label>
            <div className="col-span-10">
              <Input
                value={this.state.dataset.fullName}
                onChange={e => this.updateDatasetField("fullName", e.target.value)}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Organizer")}:</label>
            <div className="col-span-10">
              <Input
                value={this.state.dataset.organizer}
                onChange={e => this.updateDatasetField("organizer", e.target.value)}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Location")}:</label>
            <div className="col-span-10">
              <Input
                value={this.state.dataset.location}
                onChange={e => this.updateDatasetField("location", e.target.value)}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Address")}:</label>
            <div className="col-span-10">
              <Input
                value={this.state.dataset.address}
                onChange={e => this.updateDatasetField("address", e.target.value)}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("general:Status")}:</label>
            <div className="col-span-10">
              <Select value={this.state.dataset.status} onValueChange={value => this.updateDatasetField("status", value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="Public">Public (Everyone can see it)</SelectItem>
                  <SelectItem value="Hidden">Hidden (Only yourself can see it)</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Carousels")}:</label>
            <div className="col-span-10">
              <Input
                value={this.state.dataset.carousels.join(", ")}
                placeholder="Please input (comma separated)"
                onChange={e => this.updateDatasetField("carousels", e.target.value.split(",").map(s => s.trim()).filter(s => s))}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Introduction text")}:</label>
            <div className="col-span-10">
              <Input
                value={this.state.dataset.introText}
                onChange={e => this.updateDatasetField("introText", e.target.value)}
              />
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Default item")}:</label>
            <div className="col-span-10">
              <Select value={this.state.dataset.defaultItem} onValueChange={value => this.updateDatasetField("defaultItem", value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {this.state.dataset.treeItems
                    .filter(treeItem => treeItem.children.length === 0)
                    .map((treeItem) => (
                      <SelectItem key={treeItem.title} value={treeItem.title}>
                        {`${treeItem.title} | ${treeItem.titleEn}`}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-12 gap-4 items-center">
            <label className="col-span-2">{i18next.t("dataset:Language")}:</label>
            <div className="col-span-10">
              <Select value={this.state.dataset.language} onValueChange={value => this.updateDatasetField("language", value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="zh">zh</SelectItem>
                  <SelectItem value="en">en</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  submitDatasetEdit() {
    const dataset = Setting.deepCopy(this.state.dataset);
    DatasetBackend.updateDataset(this.state.dataset.owner, this.state.datasetName, dataset)
      .then((res) => {
        if (res) {
          Setting.showMessage("success", "Successfully saved");
          this.setState({
            datasetName: this.state.dataset.name,
          });
          this.props.history.push(`/datasets/${this.state.dataset.name}`);
        } else {
          Setting.showMessage("error", "failed to save: server side failure");
          this.updateDatasetField("name", this.state.datasetName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `failed to save: ${error}`);
      });
  }

  render() {
    return (
      <div className="w-full px-4 py-4">
        <div className="max-w-[95%] mx-auto">
          {this.state.dataset !== null ? this.renderDataset() : null}
        </div>
        <div className="max-w-[95%] mx-auto mt-4 ml-3">
          <Button size="lg" onClick={this.submitDatasetEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      </div>
    );
  }
}

export default DatasetEditPage;
