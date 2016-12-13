var API_URL = '/api';
var speak;

document.addEventListener("DOMContentLoaded", function(event) {
  speak = document.querySelector('.speak');
  speak.addEventListener('mousedown', Speaking.start);
  speak.addEventListener('touchstart', Speaking.start);

  speak.addEventListener('mouseup', Speaking.finish);
  speak.addEventListener('touchend', Speaking.finish);
  speak.addEventListener('touchcancel', Speaking.finish);
});

var AudioContext = window.AudioContext || window.webkitAudioContext;
navigator.getUserMedia = navigator.getUserMedia ||
                         navigator.webkitGetUserMedia ||
                         navigator.mozGetUserMedia ||
                         navigator.msGetUserMedia;

var recorder;
var audioContext;

function createAudioContext() {
  try {
    audioContext = new AudioContext();
  } catch (e) {
    alert('Web audio is not supported in this browser.');
    throw new Error('Web audio is not supported in this browser');
  }
};

function createRecorder() {
  navigator.getUserMedia(
    { audio: true },
    function(stream) {
      var input = audioContext.createMediaStreamSource(stream);
      recorder = new Recorder(input);
    }, function() {
      alert('This app is unable to work without microphone access.');
    }
  );
};

function loadingAnimation() {
  speak.classList.remove('speak__loading')
  speak.classList.add('speak__waiting')
};

var Speaking = {
  start: function(e) {
    e.returnValue = false;

    speak.classList.add('speak__listen');
    recorder.record();
  },
  finish: function() {
    speak.classList.remove('speak__listen');
    speak.classList.add('speak__loading');

    recorder.stop();
    recorder.exportWAV(function(audio) {
      audio.lastModifiedDate = new Date();
      audio.name = 'file';

      var data = new FormData();
      data.append('file', audio);

      // Add animation end transition
      speak.addEventListener('animationend', loadingAnimation, false);

      $.ajax({
        url: `${API_URL}/speech`,
        data,
        cache: false,
        contentType: false,
        processData: false,
        type: 'POST',
        complete: function() {
          recorder.clear();
          speak.classList.remove('speak__waiting');
          speak.removeEventListener('animationend', loadingAnimation, false);
        }
      });
    });
  },
};

createAudioContext();
createRecorder();
