/*
This may help https://bl.ocks.org/mbostock/7607999
*/
var svg = d3.select("svg"),
    width = +svg.attr("width"),
    height = +svg.attr("height"),
    color = d3.scaleOrdinal(d3.schemeCategory10);

var nodes = [],
    links = [],
    edges = [];

// Map from node id to its data.
var nodeInfo = {},
    edgeInfo = {};

var simulation = d3.forceSimulation(nodes)
    .force("charge", d3.forceManyBody().strength(-1000))
    .force("link", d3.forceLink(links).distance(200))
    .force("x", d3.forceX())
    .force("y", d3.forceY())
    .alphaTarget(1)
    .on("tick", ticked);

var g = svg.append("g").attr("transform", "translate(" + width / 2 + "," + height / 2 + ")"),
    link = g.append("g").attr("stroke", "#000").attr("stroke-width", 1.5).selectAll(".link"),
    node = g.append("g").attr("stroke", "#fff").attr("stroke-width", 1.5).selectAll(".node");

function parseInfoResp(data) {
  let v = JSON.parse(data);
  let pb = JSON.parse(v.Buckets)
  let pblen = pb.length;
  for (let i = 0; i<pblen; i++) {
    pb[i] = JSON.parse(pb[i]);
  }
  return {
    ID: v.ID,
    Port: v.Port,
    Buckets: pb,
    Storage: JSON.parse(v.Storage)
  };
}

$.ajax({
  type: "GET",
  url: "/sarga/api/",
  success: function(data, status, jqXHR) {
    nodes = [];
    links = [];

    // ID to Address, Buckets and Storage map.
    nodeInfo = {};

    // ID to Node object map. Contains address as well.
    idToNodeMap = {};

    edgeInfo = {};

    processNodeInfo(data);

    restart();
  }
});

function processNodeInfo(data) {
  let resp = parseInfoResp(data);
  let rootNode = {"id": resp.ID};

  nodeInfo[resp.ID] = resp;
  idToNodeMap[resp.ID] = rootNode;

  let bucket_count = resp.Buckets.length;
  for (let i=0; i<bucket_count; i++) {
    for (let neighbor_id in resp.Buckets[i]) {
      if (resp.Buckets[i].hasOwnProperty(neighbor_id)) {
        let neighbor = {
          "id": neighbor_id,
          "address": resp.Buckets[i][neighbor_id]
        };
        edges.push(rootNode.id+"-"+neighbor.id);
        idToNodeMap[neighbor_id] = neighbor;
      }
    }
  }
}

function unique(arr) {
  var u = {}, a = [];
  for(var i = 0, l = arr.length; i < l; ++i){
    if(!u.hasOwnProperty(arr[i])) {
      a.push(arr[i]);
      u[arr[i]] = 1;
    }
  }
  return a;
}

function restart() {
  nodes = [];
  links = [];
  for (let node_id in idToNodeMap) {
    if (idToNodeMap.hasOwnProperty(node_id)) {
      nodes.push(idToNodeMap[node_id]);
    }
  }

  edges = unique(edges);
  for (let i=0; i<edges.length; i++) {
    let kk = edges[i].split("-");
    links.push({source: idToNodeMap[kk[0]], target: idToNodeMap[kk[1]]});
  }

  // Apply the general update pattern to the nodes.
  node = node.data(nodes, function(d) { return d.id; });
  node.exit().remove();
  node = node
    .enter()
    .append("circle")
    .attr("fill", function(d) { return color(d.id); })
    .attr("r", 8)
    .merge(node)
    .text(function(d) { return d.id; })
    .on("click", mouseclicked)
    .on("mouseover", mouseovered)
    .on("mouseout", mouseouted);

  // Apply the general update pattern to the links.
  link = link.data(links, function(d) { return d.source.id + "-" + d.target.id; });
  link.exit().remove();
  link = link.enter().append("line").merge(link);

  // Update and restart the simulation.
  simulation.nodes(nodes);
  simulation.force("link").links(links);
  simulation.alpha(1).restart();
}

function ticked() {
  node.attr("cx", function(d) { return d.x; })
      .attr("cy", function(d) { return d.y; })

  link.attr("x1", function(d) { return d.source.x; })
      .attr("y1", function(d) { return d.source.y; })
      .attr("x2", function(d) { return d.target.x; })
      .attr("y2", function(d) { return d.target.y; });
}

function mouseclicked(d) {
  let clicked = idToNodeMap[d.id];
  $.ajax({
    type: "GET",
    url: "http://"+idToNodeMap[d.id].address+"/info",
    success: function(data, status, jqXHR) {
      processNodeInfo(data);
      restart();
    }
  });
}

function mouseovered(d) {
  console.log(JSON.stringify(d));
}

function mouseouted(d) {
}

// d3.timeout(function() {
//   links.push({source: a, target: b}); // Add a-b.
//   links.push({source: b, target: c}); // Add b-c.
//   links.push({source: c, target: a}); // Add c-a.
//   restart();
// }, 1000);

// d3.interval(function() {
//   nodes.pop(); // Remove c.
//   links.pop(); // Remove c-a.
//   links.pop(); // Remove b-c.
//   restart();
// }, 2000, d3.now());

// d3.interval(function() {
//   nodes.push(c); // Re-add c.
//   links.push({source: b, target: c}); // Re-add b-c.
//   links.push({source: c, target: a}); // Re-add c-a.
//   restart();
// }, 2000, d3.now() + 1000);
