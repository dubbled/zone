import Konva from 'konva';

console.log('hello!');

const graph = document.getElementById('graph');
const startButton = document.querySelector('button');

const ctx = CanvasCaptureMediaStreamTrack.getContext('2d')


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

function generateNode() {
  return new Konva.Circle({
    x: WIDTH * Math.random(),
    y: HEIGHT * Math.random(),
    radius: 50,
    fill: 'red',
    stroke: 'black',
  });
}

for (let i = 0; i < NUMBER; i++) {
  layer.add(generateNode());
}
