var directory = angular.module('directory', [])

directory.controller('DirectoryListing', function ($scope, $http) {
    $scope.path = document.location.pathname
    if (!$scope.path.match(/\/$/)) {
        $scope.path += '/'
    }

    $scope.files = []

    var socket = new WebSocket('ws://'+document.location.host+'/_srv/api')

    // Send the path
    socket.onopen = function () {
        socket.send($scope.path)
    }

    // Listen for updates
    socket.onmessage = function (event) {
        $scope.files = JSON.parse(event.data)
        $scope.$apply()
    }
})
