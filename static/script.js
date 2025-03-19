const postsContainer = document.getElementById("posts");
const liveList = document.getElementById("live-list");
let loadMap = {};

let maxFetchedId = 0;
let storyIds = [];
let currentIndex = 0;
const storiesPerPage = 10;
let pollPage = 0;

//let isLoading = false;
let isThrottled = false;

// Fetch stories
async function fetchStories(url) {
    currentIndex = 0;
    postsContainer.innerHTML = "";
    const response = await fetch(url);
    storyIds = await response.json();
    storyIds.sort((a, b) => b - a)  // sort stories newest first, as requested

    loadStories();

    // Load more if all fits in window so scrolling events can take place
    if (document.body.offsetHeight <= window.innerHeight) {
        loadMoreStories();  // no throttling, otherwise may not happen
    }
}

// Fetch details for a single story
async function fetchStory(id) {
    if (loadMap[id] === undefined) {
        const response = await fetch(
            `https://hacker-news.firebaseio.com/v0/item/${id}.json`
        );
        loadMap[id] = await response.json();
        return loadMap[id];
    } else {
        return loadMap[id];
    }
}

// Load set of stories
async function loadStories() {
    let endIndex = currentIndex + storiesPerPage

    // Few stories to get
    if (endIndex > storyIds.length) {
        endIndex = storyIds.length // end index is exclusive in slice()
    }

    const stories = await Promise.all(
        storyIds.slice(currentIndex, endIndex).map(fetchStory)
    );

    stories.forEach((story) => {
        if (!story.deleted && !story.dead) {
            addStoryToPage(story);
        }
    });
}

function throttle(mainFunction, delay) {
    let timerFlag = null;

    // return a throttled function
    return (...args) => {
        if (timerFlag === null) {
            mainFunction(...args); // execute the main function
            timerFlag = setTimeout(() => {
                timerFlag = null; // set a timer to clear the timerFlag
            }, delay);
        }
    };
}

async function loadMoreStories() {
    // No more stories to get
    if (currentIndex + storiesPerPage > storyIds.length) {
        return
    }

    // if poll is open and current index is divisible by 20 (algolia page size): get new elements to id list 
    if (document.getElementById('pollButton').classList.contains("activeNavigation")) {
        if (currentIndex % 20 == 0) {
            pollPage++;
            const response = await fetch(
                `https://hn.algolia.com/api/v1/search_by_date?query=&tags=poll&page=${pollPage}`
            );
            const polls = await response.json();
            polls.hits.forEach((hit) => {
                storyIds.push(hit.objectID);
            });
        }
    } else {
        pollPage = 0;
    }

    currentIndex += storiesPerPage;
    await loadStories();
}

const throttledLoadMoreStories = throttle(loadMoreStories, 5000);

// Add scroll event listener for infinite scroll
window.addEventListener("scroll", () => {
    // Check if the user has scrolled near the bottom of the page
    if (window.innerHeight + window.scrollY >= document.body.offsetHeight - 100) {
        throttledLoadMoreStories();
    }
});

// Load set of comments
async function loadComments(commentIds, parentPanel) {
    const comments = await Promise.all(commentIds.map(fetchStory));

    const repliesDiv = document.createElement("div");

    comments.forEach(async (comment) => {
        if (comment.deleted === undefined && comment.dead === undefined) {

            const commentElement = document.createElement("div");
            commentElement.id = "comment_" + comment.id;
            commentElement.classList.add("comment")

            // Write date, author and what not
            const commentDate = new Date(comment.time * 1000);
            const inputDate = String(commentDate.getDate()).padStart(2, "0");
            const inputMonth = String(commentDate.getMonth() + 1).padStart(2, "0");
            const inputYear = commentDate.getFullYear();
            const inputHours = String(commentDate.getHours()).padStart(2, "0");
            const inputMinutes = String(commentDate.getMinutes()).padStart(2, "0");
            const inputSeconds = String(commentDate.getSeconds()).padStart(2, "0");

            commentElement.innerHTML += `${comment.text}
                                </br>
                                <small>${inputYear}-${inputMonth}-${inputDate} ${inputHours}:${inputMinutes}:${inputSeconds}</small>
                                <small>by: <b>${comment.by}</b></small>
                                <small>type: <b>${comment.type}</b></small>
                                <small>replies: <b>${comment.kids === undefined
                    ? 0
                    : comment.kids.length
                }</b></small>`;
            repliesDiv.appendChild(commentElement);

            // If kids, add a new accordion to show them
            if (comment.kids !== undefined && comment.kids.length > 0) {
                commentElement.classList.add("accordion");

                const commentPanelElement = document.createElement("div");
                commentPanelElement.className = "panel";
                repliesDiv.appendChild(commentPanelElement);

                loadAccordion(commentElement, comment, commentPanelElement);
            }
        }
    });

    //parentPanel.innerHTML = "";
    parentPanel.appendChild(repliesDiv);
}

// Display story on page
async function addStoryToPage(story) {
    const postElement = document.createElement("div");
    postElement.classList.add("post");
    // Write date, author and what not
    const storyDate = new Date(story.time * 1000);
    const inputDate = String(storyDate.getDate()).padStart(2, "0");
    const inputMonth = String(storyDate.getMonth() + 1).padStart(2, "0");
    const inputYear = storyDate.getFullYear();
    const inputHours = String(storyDate.getHours()).padStart(2, "0");
    const inputMinutes = String(storyDate.getMinutes()).padStart(2, "0");
    const inputSeconds = String(storyDate.getSeconds()).padStart(2, "0");

    const href = story.url === undefined ? "" : `href="${story.url}"`;

    let inner = ''
    href ? inner += `<a ${href} target="_blank">${story.title}</a></br>` : inner += `${story.title}</br>`;
    inner += `<small>${inputYear}-${inputMonth}-${inputDate} ${inputHours}:${inputMinutes}:${inputSeconds}</small>
              <small>by: <b>${story.by}</b></small>
              <small>type: <b>${story.type}</b></small>
              <small>score: ${story.score}</small>`;
    story.descendants ? inner += `<small> comments: ${story.descendants}</small>` : ``;
    postElement.innerHTML = inner;

    postsContainer.appendChild(postElement);

    if (story.text !== undefined || story.descendants > 0 || (story.type == 'poll' && story.parts)) {
        postElement.classList.add("accordion");

        const panelDivElement = document.createElement("div");
        panelDivElement.className = "panel";
        postsContainer.appendChild(panelDivElement);

        if (story.text !== undefined) {
            const textContentDiv = document.createElement('div')
            textContentDiv.classList.add('textcontent')
            const textDiv = document.createElement('p')
            textDiv.classList.add('actualText')            
            textDiv.innerHTML = story.text;
            textContentDiv.appendChild(textDiv)
            panelDivElement.appendChild(textContentDiv)
        }

        loadAccordion(postElement, story, panelDivElement);
    }
}

const loadAccordion = (element, data, panelDivElement) => {
    element.addEventListener("click", async function () {
        // add/remove "active" class for styling in css
        this.classList.toggle("active");

        // Toggle between hiding and showing the active panel
        if (panelDivElement.style.display === "block") {
            panelDivElement.style.display = "none";
        } else {
            panelDivElement.style.display = "block";

            // add poll parts when appropriate
            if (data.type !== undefined && data.type === 'poll' && data.parts !== undefined && Array.isArray(data.parts)) {
                const parts = await Promise.all(
                    data.parts.map(fetchStory)
                );

                let textDiv
                // see if panel already has text element where to put poll options
                if (!panelDivElement.querySelector(".textcontent")) {
                    const textContentDiv = document.createElement('div')
                    textContentDiv.classList.add('textcontent')
                    panelDivElement.appendChild(textContentDiv);
                    textDiv = textContentDiv
                } else {
                    textDiv = panelDivElement.querySelector(".textcontent")
                }

                // Remove any old poll-options
                textDiv.querySelectorAll(".poll-option").forEach(child => child.remove());

                parts.sort((a, b) => b.score - a.score); // Sort by votes descending
                parts.forEach(option => {
                    const item = document.createElement("div");
                    item.className = "poll-option";
                    item.innerHTML = `<strong>${option.score}</strong> \t ${option.text}`;
                    textDiv.appendChild(item);
                });
            }

            // Remove any old comments
            panelDivElement.querySelectorAll(".comment").forEach(child => child.remove());

            if (data.kids !== undefined && data.kids.length > 0) {
                data.kids.sort((a, b) => b - a)  // sort comments newest first, as requested
                await loadComments(data.kids, panelDivElement);
            }
        }
    });
};

// Fetch new posts every 5 seconds
async function checkForLiveUpdates() {
    const response = await fetch(
        "https://hacker-news.firebaseio.com/v0/newstories.json"
    );
    const newStories = await response.json();

    const latestStories = await Promise.all(
        newStories.slice(0, 5).map(fetchStory)
    );
    if (
        latestStories !== undefined &&
        latestStories &&
        latestStories.length > 0 &&
        latestStories[0] &&
        latestStories[0].id !== maxFetchedId
    ) {
        showToast("You have some new updates");
        maxFetchedId = latestStories[0].id;
    }

    liveList.innerHTML = ""; // Clear previous update
    latestStories.forEach((story) => {
        const listItem = document.createElement("li");

        if (story !== undefined && story !== null) {
            if (story.url !== undefined && story.url) {
                listItem.innerHTML = `<a href="${story.url}" target="_blank">${story.title}</a>`;
            } else {
                listItem.innerHTML = `${story.title}`;
            }
        } /* else {
            listItem.innerHTML = `title missing`;
        } */
        liveList.appendChild(listItem);
    });
}

// Fetch updated posts every 20 seconds
async function checkForPostUpdates() {
    const response = await fetch(
        "https://hacker-news.firebaseio.com/v0/updates.json"
    );
    const updated = await response.json();
    const updatedItems = updated["items"];

    let toastTotalMessage = [];
    if (Object.keys(loadMap).length > 0) {
        updatedItems.forEach((updatedItem) => {
            if (loadMap.hasOwnProperty(updatedItem) && loadMap[updatedItem]) {
                let toastMessage = loadMap[updatedItem].type + ' ';
                toastMessage += `id #${updatedItem}`;

                if (loadMap[updatedItem].title) {
                    toastMessage += ` "${loadMap[updatedItem].title}"`;
                }
                toastMessage += " updated";
                delete loadMap[updatedItem];

                toastTotalMessage.push(toastMessage);
            }
        });
        if (toastTotalMessage.length > 0) {
            showToast(toastTotalMessage.join("\n"));
        }
    }
}

addEventListener("DOMContentLoaded", function () {
    // Initial Load
    fetchStories("https://hacker-news.firebaseio.com/v0/topstories.json");
    setInterval(checkForLiveUpdates, 5000);
    setInterval(checkForPostUpdates, 20000);

    const deactiveAllNavigations = function () {
        const navigationButtons =
            document.getElementsByClassName("navigationButton");
        for (const navigationButton of Array.from(navigationButtons)) {
            navigationButton.classList.remove("activeNavigation");
        }
    };

    const homeButton = document.getElementById("homeButton");
    homeButton.addEventListener("click", function () {
        deactiveAllNavigations();
        homeButton.classList.add("activeNavigation");
        fetchStories("https://hacker-news.firebaseio.com/v0/topstories.json");
    });
    homeButton.classList.add("activeNavigation");

    const newestButton = document.getElementById("newestButton");
    newestButton.addEventListener("click", function () {
        deactiveAllNavigations();
        newestButton.classList.add("activeNavigation");
        fetchStories("https://hacker-news.firebaseio.com/v0/newstories.json");
    });

    const bestButton = document.getElementById("bestButton");
    bestButton.addEventListener("click", function () {
        deactiveAllNavigations();
        bestButton.classList.add("activeNavigation");
        fetchStories("https://hacker-news.firebaseio.com/v0/beststories.json");
    });

    const askButton = document.getElementById("askButton");
    askButton.addEventListener("click", function () {
        deactiveAllNavigations();
        askButton.classList.add("activeNavigation");
        fetchStories("https://hacker-news.firebaseio.com/v0/askstories.json");
    });

    const showButton = document.getElementById("showButton");
    showButton.addEventListener("click", function () {
        deactiveAllNavigations();
        showButton.classList.add("activeNavigation");
        fetchStories("https://hacker-news.firebaseio.com/v0/showstories.json");
    });

    const jobButton = document.getElementById("jobButton");
    jobButton.addEventListener("click", function () {
        deactiveAllNavigations();
        jobButton.classList.add("activeNavigation");
        fetchStories("https://hacker-news.firebaseio.com/v0/jobstories.json");
    });

    const pollButton = document.getElementById("pollButton");
    pollButton.addEventListener("click", async function () {
        deactiveAllNavigations();
        pollButton.classList.add("activeNavigation");

        // Get polls with algolia search API
        currentIndex = 0;
        postsContainer.innerHTML = "";
        const response = await fetch(
            `https://hn.algolia.com/api/v1/search_by_date?query=&tags=poll&page=${pollPage}`
        );
        const polls = await response.json();

        storyIds = [];
        polls.hits.forEach((hit) => {
            storyIds.push(hit.objectID);
        });

        loadStories();

        // Load more if all fits in window so scrolling events can take place
        if (document.body.offsetHeight <= window.innerHeight) {
            loadMoreStories();
        }
    });
});
