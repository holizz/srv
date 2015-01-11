var directory = angular.module('directory', [])

directory.controller('DirectoryListing', function ($scope, $http) {
    $scope.error = true

    $scope.path = decodeURI(document.location.pathname)
    if (!$scope.path.match(/\/$/)) {
        $scope.path += '/'
    }

    $scope.files = []

    var connect = function () {
        var socket = new WebSocket('ws://'+document.location.host+'/_srv/api')

        // Send the path
        socket.onopen = function () {
            socket.send($scope.path+"\n")
        }

        // Listen for updates
        socket.onmessage = function (event) {
            $scope.files = JSON.parse(event.data)
            $scope.error = false
            $scope.$apply()
        }

        // "uh oh"
        socket.onclose = function (event) {
            $scope.error = true
            $scope.$apply()
        }
    }

    connect()
    window.setInterval(function () {
        if ($scope.error) {
            connect()
        }
    }, 5000)

})
