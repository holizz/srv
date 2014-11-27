var directory = angular.module('directory', [])

directory.controller('DirectoryListing', function ($scope, $http) {
    $scope.error = false

    $scope.path = document.location.pathname
    if (!$scope.path.match(/\/$/)) {
        $scope.path += '/'
    }

    $scope.files = []

    var socket = new WebSocket('ws://'+document.location.host+'/_srv/api')

    // Send the path
    socket.onopen = function () {
        socket.send($scope.path+"\n")
    }

    // Listen for updates
    socket.onmessage = function (event) {
        $scope.files = JSON.parse(event.data)
        $scope.$apply()
    }

    // "uh oh"
    socket.onclose = function (event) {
        $scope.error = true
        $scope.$apply()
    }
})
