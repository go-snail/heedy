import { LTTB } from "../../../dist/downsample.mjs";

function extractor(arr) {
  if (arr.length == 1) {
    if (arr[0] == "t") {
      return (dp) => dp.t * 1000; // Timestamps are multiplied by 1000
    }
    if (arr[0] == "d") {
      return (dp) => dp.d;
    }
    if (arr[0] == "dt") {
      return (dp) => dp.dt * 1000;
    }
  }
  return (dp) => {
    for (let i = 0; i < arr.length; i++) {
      dp = dp[arr[i]];
      if (dp === undefined) {
        return null;
      }
    }
    return dp;
  };
}

function prepareDataset(qd, ds) {
  let newds = {
    ...ds,
  };

  // First, extract the desired data from the dataset

  let extractX = extractor(ds.data.x);
  let extractY = extractor(ds.data.y);
  if (ds.data.withDuration !== undefined && ds.data.widthDuration) {
    let data = qd.dataset[ds.data.series].map((dp) => ({
      t: dp.t * 1000,
      y: extractY(dp),
      dt: dp.dt * 1000,
    }));

    dataset = new Array(data.length * 3); // Start point, endpoint, and a null to break the line

    for (let i = 0; i < data.length; i++) {
      dataset[i * 3] = {
        x: data[i].t,
        y: data[i].y,
      };
      dataset[i * 3 + 1] = {
        x: data[i].t + data[i].dt,
        y: data[i].y,
      };
      dataset[i * 3 + 2] = null;
    }

    newds.data = dataset;

    // If the points are colored, only show the point at start

    if (newds.pointBackgroundColor !== undefined) {
      let pointColor = new Array(data.length * 3);

      for (let i = 0; i < data.length; i++) {
        pointColor[i * 3] = newds.pointBackgroundColor;
        pointColor[i * 3 + 1] = "transparent";
        pointColor[i * 3 + 2] = "transparent";
      }
      newds.pointBackgroundColor = pointColor;
    }
    if (newds.pointBorderColor !== undefined) {
      let pointColor = new Array(data.length * 3);

      for (let i = 0; i < data.length; i++) {
        pointColor[i * 3] = newds.pointBorderColor;
        pointColor[i * 3 + 1] = "transparent";
        pointColor[i * 3 + 2] = "transparent";
      }
      newds.pointBorderColor = pointColor;
    }
  } else {
    newds.data = qd.dataset[ds.data.series].map((dp) => ({
      x: extractX(dp),
      y: extractY(dp),
    }));
  }

  if (ds.data.downsample !== undefined && ds.data.downsample > 0) {
    // The data needs to be downsampled
    newds.data = LTTB(newds.data, ds.data.downsample);
  }

  return newds;
}

function preprocess(qd, visualization) {
  // The preprocessing step generates the actual data from the dataset according to the configuration

  // The output is a copy of the object with the datasets replaced
  let nvis = {
    ...visualization,
    config: {
      ...visualization.config,
      charts: visualization.config.charts.map((c) => ({
        ...c,
        data: {
          ...c.data,
          datasets: c.data.datasets.map((ds) => prepareDataset(qd, ds)),
        },
      })),
    },
  };

  if (nvis.config.syncX !== undefined && nvis.config.syncX) {
    // Change the X axis  minimum and maximum values to the total dataset max and min

    let totalMin = Math.min.apply(
      null,
      nvis.config.charts.map((c) =>
        Math.min.apply(
          null,
          c.data.datasets.map((ds) =>
            ds.data.length > 0 ? ds.data[0].x : Infinity
          )
        )
      )
    );
    // Max needs to be aware of possible nulls if the points are withDuration
    let totalMax = Math.max.apply(
      null,
      nvis.config.charts.map((c) =>
        Math.max.apply(
          null,
          c.data.datasets.map((ds) =>
            ds.data.length == 0
              ? -Infinity
              : ds.data[ds.data.length - 1] !== null
              ? ds.data[ds.data.length - 1].x
              : ds.data.length >= 2 && ds.data[ds.data.length - 2] !== null
              ? ds.data[ds.data.length - 2].x
              : -Infinity
          )
        )
      )
    );

    let bounds = {
      min: totalMin,
      max: totalMax,
    };

    // Now set this value for all x axes
    nvis.config.charts = nvis.config.charts.map((c) => ({
      ...c,
      options: {
        ...c.options,
        scales: {
          ...c.options.scales,
          xAxes: c.options.scales.xAxes.map((ax) => ({ ...ax, ticks: bounds })),
        },
      },
    }));
  }

  return nvis;
}

export default preprocess;