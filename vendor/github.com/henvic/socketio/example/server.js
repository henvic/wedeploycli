'use strict'

let app = require('http').createServer(handler)
let io = require('socket.io')(app)

app.listen(3000)

console.log('Listening at http://localhost:3000/')

function handler (req, res) {
  res.writeHead(200)
  res.end('Testing server for http://github.com/henvic/socketio example.')
}

let vehicles = ['Falcon', 'airplane', 'balloon', 'drone']
let airports = ['JFK', 'KEF', 'ATL', 'MIA', 'DAO', 'FCO']
let people = ['Stephen', 'Albert', 'Thomas', 'George', 'Michael', 'Linus']
let rooms = ['King', 'Queen', 'Presidential suite']

class AirNotices {
  getStatusFunc (socket) {
    return () => this.getStatus(socket)
  }

  getStatus (socket) {
    let vehicle = vehicles[Math.floor(Math.random() * vehicles.length)]
    let index1 = Math.floor(Math.random() * airports.length)
    let index2 = Math.floor(Math.random() * airports.length)

    if (index1 === index2) {
      socket.emit('skip', vehicle)
      return
    }

    socket.emit('flight', vehicle, {
      From: airports[index1],
      To: airports[index2]
    })
  }
}

let an = new AirNotices()

io.on('connection', function (socket) {
  console.log('Connecting %s.', socket.id)

  let anInterval = setInterval(an.getStatusFunc(socket), 2500)

  let scheduledGoodbyeTimeout = setTimeout(function () {
    console.log('Sending goodbye message to client.')
    socket.emit('goodbye')
  }, 20000)

  socket.on('book_hotel_for_tonight', (location, fn) => {
    console.log('Hotel room booked at %s Airport Hotel.', location)
    let indexPerson = Math.floor(Math.random() * people.length)
    let indexRoom = Math.floor(Math.random() * rooms.length)

    fn({
      Name: people[indexPerson],
      Location: location,
      Room: rooms[indexRoom],
      Price: Math.ceil(Math.random() * 1000)
    })
  })

  socket.on('find_tickets', (route) => {
    console.log('Quote for tickets from %s to %s requested.', route.From, route.To)
  })

  socket.on('book_hotel_for_tonight', (location, cb) => {
    console.log('Booking hotel room in %s.', location)
          cb('Hello!') // eslint-disable-line
  })

  socket.on('error', (err) => {
    console.error(err)
  })

  socket.on('disconnect', () => {
    console.log('Disconnecting %s.', socket.id)
    clearInterval(anInterval)
    clearInterval(scheduledGoodbyeTimeout)
  })
})
