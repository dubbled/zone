<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Konva Example</title>
  <script src="https://unpkg.com/konva@8/konva.min.js"></script>
  <style>
    .label {
      font-size: 20px;
      fill: black; // Change the font color to black
      text-anchor: middle;
    }
  </style>
</head>
<body>
  <div id="container"></div>
  <script>
const WIDTH = 3000;
const HEIGHT = 3000;
const NUMBER = 200;

const stage = new Konva.Stage({
  container: 'container',
  width: WIDTH,
  height: HEIGHT,
});

const layer = new Konva.Layer();
stage.add(layer);

    function generateNode(x, y, color, label) {
      const circle = new Konva.Circle({
    x: x + 50,
    y: y + 50,
        radius: 10,
        fill: color,
    stroke: 'black',
    draggable: true
  });

      const text = new Konva.Text({
    x: x + 50,
        y: y - 30,
        text: label,
        fontSize: 20,
    fill: 'black', // Change the font color to black
    textAlign: 'center'
      });

  circle.on('dragmove', function () {
    text.x(circle.x() + 50);
    text.y(circle.y() - 30);
  });

      return [circle, text];
}

    const colors = ['red', 'green', 'blue'];
    const labels = ['Manpower', 'Ore', 'Food'];

    for (let i = 0; i < NUMBER; i++) {
      const x = Math.floor(Math.random() * (WIDTH / 100)) * 100;
      const y = Math.floor(Math.random() * (HEIGHT / 100)) * 100;

      const index = Math.floor(Math.random() * colors.length);
      const [circle, text] = generateNode(x, y, colors[index], labels[index]);

      let originalPosition = { x: circle.x(), y: circle.y() };

      circle.on('dragend', function (e) {
        const gridX = Math.round(e.target.x() / 100) * 100;
        const gridY = Math.round(e.target.y() / 100) * 100;

        e.target.x(gridX + 50);
        e.target.y(gridY + 50);

        originalPosition = { x: gridX + 50, y: gridY + 50 };
      });

      layer.add(circle);
      layer.add(text);
    }

    function drawGrid() {
      const gridLayer = new Konva.Layer();
      for (let i = 0; i <= HEIGHT / 100; i++) {
        gridLayer.add(new Konva.Line({
          points: [0, i * 100, WIDTH, i * 100],
          stroke: 'gray',
      strokeWidth: 1
        }));
    }
  for (let j = 0; j <= WIDTH / 100; j++) {
    gridLayer.add(new Konva.Line({
      points: [j * 100, 0, j * 100, HEIGHT],
      stroke: 'gray',
      strokeWidth: 1
    }));
  }
  stage.add(gridLayer);
}

    drawGrid();
  </script>
</body>
</html>

