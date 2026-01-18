import React from "react";
import {Link} from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import * as Setting from "./Setting";
import * as DatasetBackend from "./backend/DatasetBackend";
import i18next from "i18next";
import { format } from "date-fns";

class DatasetListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      datasets: null,
      deleteDialogOpen: false,
      datasetToDelete: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getDatasets();
  }

  getDatasets() {
    DatasetBackend.getDatasets(this.props.account.name)
      .then((res) => {
        this.setState({
          datasets: res,
        });
      });
  }

  newDataset() {
    const now = new Date();
    const dateStr = format(now, "yyyy-MM-dd");
    return {
      owner: this.props.account.name,
      name: `dataset_${this.state.datasets.length}`,
      createdTime: now.toISOString(),
      startDate: dateStr,
      endDate: dateStr,
      fullName: `Dataset ${this.state.datasets.length}`,
      organizer: "Casbin",
      location: "Shanghai, China",
      address: "3663 Zhongshan Road North",
      status: "Public",
      language: "zh",
      carousels: [],
      introText: "Introduction..",
      defaultItem: "Home",
      treeItems: [{key: "Home", title: "首页", titleEn: "Home", content: "内容", contentEn: "Content", children: []}],
    };
  }

  addDataset() {
    const newDataset = this.newDataset();
    DatasetBackend.addDataset(newDataset)
      .then((res) => {
        Setting.showMessage("success", "Dataset added successfully");
        this.setState({
          datasets: Setting.prependRow(this.state.datasets, newDataset),
        });
      }
      )
      .catch(error => {
        Setting.showMessage("error", `Dataset failed to add: ${error}`);
      });
  }

  deleteDataset(i) {
    DatasetBackend.deleteDataset(this.state.datasets[i])
      .then((res) => {
        Setting.showMessage("success", "Dataset deleted successfully");
        this.setState({
          datasets: Setting.deleteRow(this.state.datasets, i),
          deleteDialogOpen: false,
          datasetToDelete: null,
        });
      }
      )
      .catch(error => {
        Setting.showMessage("error", `Dataset failed to delete: ${error}`);
      });
  }

  renderTable(datasets) {
    if (datasets === null) {
      return <div className="flex justify-center items-center h-64">Loading...</div>;
    }

    return (
      <div className="border rounded-lg">
        <div className="p-4 border-b bg-white flex items-center justify-between">
          <h2 className="text-lg font-semibold">{i18next.t("general:Datasets")}</h2>
          <Button size="sm" onClick={this.addDataset.bind(this)}>{i18next.t("general:Add")}</Button>
        </div>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{i18next.t("general:Name")}</TableHead>
              <TableHead>{i18next.t("dataset:Start date")}</TableHead>
              <TableHead>{i18next.t("dataset:End date")}</TableHead>
              <TableHead>{i18next.t("dataset:Full name")}</TableHead>
              <TableHead>{i18next.t("dataset:Organizer")}</TableHead>
              <TableHead>{i18next.t("dataset:Location")}</TableHead>
              <TableHead>{i18next.t("dataset:Address")}</TableHead>
              <TableHead>{i18next.t("general:Status")}</TableHead>
              <TableHead>{i18next.t("general:Action")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {datasets.map((record, index) => (
              <TableRow key={record.name}>
                <TableCell>
                  <Link to={`/datasets/${record.name}`} className="text-primary hover:underline">
                    {record.name}
                  </Link>
                </TableCell>
                <TableCell>{Setting.getFormattedDate(record.startDate)}</TableCell>
                <TableCell>{Setting.getFormattedDate(record.endDate)}</TableCell>
                <TableCell>{record.fullName}</TableCell>
                <TableCell>{record.organizer}</TableCell>
                <TableCell>{record.location}</TableCell>
                <TableCell>{record.address}</TableCell>
                <TableCell>{record.status}</TableCell>
                <TableCell>
                  <div className="flex gap-2">
                    <Button size="sm" onClick={() => this.props.history.push(`/datasets/${record.name}`)}>
                      {i18next.t("general:Edit")}
                    </Button>
                    <ConfirmDialog
                      open={this.state.deleteDialogOpen && this.state.datasetToDelete === index}
                      onOpenChange={(open) => this.setState({ deleteDialogOpen: open, datasetToDelete: open ? index : null })}
                      title="Delete Dataset"
                      description={`Sure to delete dataset: ${record.name}?`}
                      onConfirm={() => this.deleteDataset(index)}
                    >
                      <Button size="sm" variant="destructive">
                        {i18next.t("general:Delete")}
                      </Button>
                    </ConfirmDialog>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  }

  render() {
    return (
      <div className="w-full px-4 py-4">
        <div className="max-w-[95%] mx-auto">
          {this.renderTable(this.state.datasets)}
        </div>
      </div>
    );
  }
      </div>
    );
  }
}

export default DatasetListPage;
